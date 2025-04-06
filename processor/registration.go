package processor

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wa-serv/repository"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// ProcessRegistration handles registration commands in the format "REG#Name#Address"
func ProcessRegistration(client *whatsmeow.Client, db *sql.DB, message string, senderJID string) error {
	// Check if the message starts with REG
	if !strings.HasPrefix(strings.ToUpper(message), "REG#") {
		return nil // Not a registration command
	}

	// Split the message by "#"
	parts := strings.Split(message, "#")
	if len(parts) != 3 {
		sendResponse(client, senderJID, "Format salah! Gunakan: REG#Nama#Alamat")
		return fmt.Errorf("invalid registration format")
	}

	// Extract the name and address
	name := strings.TrimSpace(parts[1])
	address := strings.TrimSpace(parts[2])

	// Validate inputs
	if name == "" || address == "" {
		sendResponse(client, senderJID, "Nama dan Alamat tidak boleh kosong!")
		return fmt.Errorf("empty name or address")
	}

	// Extract phone number from JID format (e.g., "123456789@s.whatsapp.net")
	phoneNumber := extractPhoneNumber(senderJID)

	// Check if user is already registered
	isRegistered, err := repository.IsMemberRegistered(db, phoneNumber)
	if err != nil {
		sendResponse(client, senderJID, "Terjadi kesalahan saat memeriksa registrasi.")
		return err
	}

	if isRegistered {
		sendResponse(client, senderJID, "Anda sudah terdaftar sebelumnya!")
		return nil
	}

	// Register the member
	err = repository.RegisterMember(db, name, address, phoneNumber)
	if err != nil {
		sendResponse(client, senderJID, "Gagal mendaftarkan anggota. Silakan coba lagi.")
		return err
	}

	// Send success message
	successMsg := fmt.Sprintf("✅ Registrasi Berhasil!\n\nNama: %s\nAlamat: %s\n\nTerima kasih telah mendaftar!", name, address)
	sendResponse(client, senderJID, successMsg)

	return nil
}

// extractPhoneNumber extracts the phone number from a WhatsApp JID
func extractPhoneNumber(jid string) string {
	parts := strings.Split(jid, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return jid
}

// sendResponse sends a WhatsApp message response
func sendResponse(client *whatsmeow.Client, to string, text string) {
	msg := &waProto.Message{
		Conversation: proto.String(text),
	}

	// Parse JID using the correct function
	jid, err := types.ParseJID(to)
	if err != nil {
		fmt.Printf("Error parsing JID: %v\n", err)
		return
	}

	_, err = client.SendMessage(context.Background(), jid, msg)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}
