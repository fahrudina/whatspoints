package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wa-serv/processor"
	"github.com/wa-serv/s3uploader"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func HandleMessageEvent(v *events.Message, db *sql.DB, client *whatsmeow.Client) {
	var msgText string
	if v.Message.GetExtendedTextMessage().GetText() != "" {
		msgText = v.Message.GetExtendedTextMessage().GetText()
	} else {
		msgText = v.Message.GetConversation()
	}

	msgText = strings.ToLower(strings.TrimSpace(msgText)) // Make the message case-insensitive
	fmt.Printf("Received message from %s: %s\n", v.Info.Sender.String(), msgText)

	if v.Message.GetImageMessage() != nil {
		handleMediaMessage(v, db, client)
	} else if msgText == "menu" {
		handleMenu(v, client)
	} else if msgText == "1" {
		handleCheckPoints(v, db, client)
	} else if msgText == "2" {
		handleRedeemInstructions(v, client)
	} else if msgText == "3" {
		handlePointRewards(v, client)
	} else if isUpsertPointsCommand(msgText) {
		handleUpsertPoints(v, db, client, msgText)
	} else if isRedeemPointsCommand(msgText) {
		handleRedeemPoints(v, db, client, msgText)
	} else {
		err := processor.ProcessRegistration(client, db, msgText, v.Info.Sender.String())
		if err != nil {
			fmt.Printf("Registration processing error: %v\n", err)
		}

		if msgText == "ping" {
			replyToMessage(v, client)
		} else if msgText == "help" {
			sendHelpMessage(v, client)
		}
	}
}

func handleMenu(evt *events.Message, client *whatsmeow.Client) {
	menuText := `📋 *Menu* 📋

Balas dengan angka pilihan Anda:
1️⃣ Cek Total Poin yang Anda miliki.
2️⃣ Tukarkan Poin.
3️⃣ Lihat Hadiah Poin.`
	msg := &waProto.Message{
		Conversation: proto.String(menuText),
	}
	_, err := client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Gagal mengirim menu: %v\n", err)
	}
}

func handleCheckPoints(evt *events.Message, db *sql.DB, client *whatsmeow.Client) {
	phoneNumber := evt.Info.Sender.String()
	memberID, err := processor.GetMemberIDByPhoneNumber(db, phoneNumber)
	if err != nil {
		sendErrorMessage(evt, client, "Gagal mengambil data poin Anda. Silakan coba lagi nanti.")
		return
	}

	currentPoints, err := processor.GetCurrentPoints(db, memberID)
	if err != nil {
		if err.Error() == fmt.Sprintf("no points record found for member ID: %d", memberID) {
			sendErrorMessage(evt, client, "Anda tidak memiliki catatan poin.")
		} else {
			sendErrorMessage(evt, client, "Gagal mengambil data poin Anda. Silakan coba lagi nanti.")
		}
		return
	}

	msg := &waProto.Message{
		Conversation: proto.String(fmt.Sprintf("Poin Anda saat ini: %d", currentPoints)),
	}
	_, err = client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Gagal mengirim poin: %v\n", err)
	}
}

func handleRedeemInstructions(evt *events.Message, client *whatsmeow.Client) {
	instructions := `Untuk menukarkan poin Anda, gunakan format berikut:
RED#<jumlah poin yang ingin ditukarkan>
Contoh: RED#50`
	msg := &waProto.Message{
		Conversation: proto.String(instructions),
	}
	_, err := client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Gagal mengirim instruksi penukaran poin: %v\n", err)
	}
}

func handleMediaMessage(evt *events.Message, db *sql.DB, client *whatsmeow.Client) {
	imageMessage := evt.Message.GetImageMessage()
	if imageMessage != nil {
		fmt.Printf("Received an image message from %s\n", evt.Info.Sender.String())

		data, err := client.Download(imageMessage)
		if err != nil {
			fmt.Printf("Failed to download image: %v\n", err)
			return
		}

		memberID, err := processor.GetMemberIDByPhoneNumber(db, evt.Info.Sender.String())
		if err != nil {
			fmt.Printf("Failed to retrieve member ID: %v\n", err)
			return
		}

		imageURL, err := s3uploader.UploadToS3(data)
		if err != nil {
			fmt.Printf("Failed to upload image to S3: %v\n", err)
			return
		}

		err = processor.SaveImageURL(db, memberID, imageURL)
		if err != nil {
			fmt.Printf("Failed to save image URL to database: %v\n", err)
			return
		}

		msg := &waProto.Message{
			Conversation: proto.String("Image received and saved successfully."),
		}
		_, err = client.SendMessage(context.Background(), evt.Info.Sender, msg)
		if err != nil {
			fmt.Printf("Error sending acknowledgment: %v\n", err)
		}
	}
}

func handleUpsertPoints(evt *events.Message, db *sql.DB, client *whatsmeow.Client, msgText string) {
	err := processor.ProcessUpsertPoints(db, evt.Info.Sender.String(), msgText)
	if err != nil {
		fmt.Printf("Failed to process upsert points: %v\n", err)
		sendErrorMessage(evt, client, err.Error())
		return
	}

	msg := &waProto.Message{
		Conversation: proto.String("Points updated successfully."),
	}
	_, err = client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Error sending acknowledgment: %v\n", err)
	}
}

func handleRedeemPoints(evt *events.Message, db *sql.DB, client *whatsmeow.Client, msgText string) {
	parts := strings.Split(msgText, "#")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "red") {
		sendErrorMessage(evt, client, "Format penukaran poin tidak valid. Gunakan format RED#<jumlah_poin>")
		return
	}

	pointsToRedeem, err := strconv.Atoi(parts[1])
	if err != nil || pointsToRedeem <= 0 {
		sendErrorMessage(evt, client, "Jumlah poin tidak valid. Gunakan angka positif.")
		return
	}

	reward, err := processor.RedeemPoints(db, evt.Info.Sender.String(), pointsToRedeem)
	if err != nil {
		if err == processor.ErrMinimumPoints {
			sendErrorMessage(evt, client, "Minimal poin untuk penukaran adalah 20.")
		} else if err == processor.ErrInvalidPoints {
			sendErrorMessage(evt, client, "Jumlah poin tidak valid untuk penukaran. Silakan pilih hadiah yang tersedia. Kirim '3' untuk melihat hadiah.")
		} else if err == processor.ErrInsufficientPoints {
			sendErrorMessage(evt, client, "Poin Anda tidak mencukupi untuk penukaran. Kirim '1' untuk cek poin Anda.")
		} else {
			fmt.Printf("Gagal menukarkan poin: %v\n", err)
			sendErrorMessage(evt, client, "Terjadi kesalahan saat memproses permintaan Anda.")
		}
		return
	}

	// Retrieve the user's ID and name in a single query
	_, memberName, err := processor.GetMemberDetailsByPhoneNumber(db, evt.Info.Sender.String())
	if err != nil {
		sendErrorMessage(evt, client, "Gagal mengambil data member. Silakan coba lagi nanti.")
		return
	}

	// Prepare the success message
	redeemID := fmt.Sprintf("RL-%s-#%d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
	successMessage := fmt.Sprintf(`🎉 *Penukaran Poin Berhasil!* 🎉
Terima kasih sudah setia bersama *Ruang Laundry*.

📌 *Detail Redeem:*

*Nama*: %s
*Poin Ditukar*: %d poin
*Hadiah*: %s

🔐 *ID Redeem:* %s
_(Harap simpan ID ini sebagai bukti klaim hadiah)_

📦 Hadiah akan segera kami proses dalam waktu *1–3 hari kerja*.
Jika ada kendala atau pertanyaan, silakan hubungi admin melalui WhatsApp.`, memberName, pointsToRedeem, reward, redeemID)

	// Send the success message
	msg := &waProto.Message{
		Conversation: proto.String(successMessage),
	}
	_, err = client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Gagal mengirim pesan konfirmasi penukaran: %v\n", err)
	}
}

func isUpsertPointsCommand(msgText string) bool {
	return len(msgText) > 6 && strings.EqualFold(msgText[:6], "input#")
}

func isRedeemPointsCommand(msgText string) bool {
	return len(msgText) > 4 && strings.EqualFold(msgText[:4], "red#")
}

func replyToMessage(evt *events.Message, client *whatsmeow.Client) {
	msg := &waProto.Message{
		Conversation: proto.String("pong"),
	}
	_, err := client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func sendHelpMessage(evt *events.Message, client *whatsmeow.Client) {
	helpText := `Available commands:
- ping: Bot responds with "pong"
- help: Shows this help message`

	msg := &waProto.Message{
		Conversation: proto.String(helpText),
	}
	_, err := client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Error sending help message: %v\n", err)
	}
}

func sendErrorMessage(evt *events.Message, client *whatsmeow.Client, errorMsg string) {
	msg := &waProto.Message{
		Conversation: proto.String(fmt.Sprintf("Error: %s", errorMsg)),
	}
	_, err := client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Error sending error message: %v\n", err)
	}
}

func handlePointRewards(evt *events.Message, client *whatsmeow.Client) {
	rewardsText := `🎁 *Hadiah Poin* 🎁

Poin dapat ditukarkan dengan layanan gratis, produk premium, atau hadiah menarik:

🧺 20 poin = Gratis cuci 2 kg.

🧺 50 poin = Gratis cuci 5 kg.

🌸 100 poin = Pewangi premium atau gratis cuci 10 kg.

🎟️ 150 poin = Voucher belanja Rp75.000.

💵 200 poin = Uang tunai Rp100.000 (dapat ditransfer ke rekening atau e-wallet).`
	msg := &waProto.Message{
		Conversation: proto.String(rewardsText),
	}
	_, err := client.SendMessage(context.Background(), evt.Info.Sender, msg)
	if err != nil {
		fmt.Printf("Gagal mengirim hadiah poin: %v\n", err)
	}
}
