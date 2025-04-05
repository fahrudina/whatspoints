# wa-serv: WhatsApp Receipt Point Calculation

This repository provides a sample implementation using the [Whatsmeow](https://github.com/tulir/whatsmeow) package for a use case where users can upload receipts via WhatsApp, and the system calculates points based on the receipt data.

## Features

- Integration with WhatsApp using the Whatsmeow package.
- Receipt upload handling via WhatsApp messages.
- Point calculation logic based on uploaded receipt data.
- Configurable database connection for storing and retrieving data.

## Installation

To use this project, clone the repository and install the required dependencies:

```bash
git clone https://github.com/wa-serv/wa-serv.git
cd wa-serv
go mod tidy
```

## Usage

### Running Locally

1. Set up the required environment variables for database configuration (see below).
2. Run the application:

```bash
go run main.go
```

3. Connect your WhatsApp account using the Whatsmeow package to start receiving messages.

### Running with Docker

1. Build the Docker image:

```bash
docker build -t wa-serv .
```

2. Run the container:

```bash
docker run -d --name wa-serv -p 8080:8080 \
  -e DB_HOST=your_db_host \
  -e DB_PORT=your_db_port \
  -e DB_USERNAME=your_db_username \
  -e DB_PASSWORD=your_db_password \
  -e DB_NAME=your_db_name \
  wa-serv
```

Replace `your_db_host`, `your_db_port`, `your_db_username`, `your_db_password`, and `your_db_name` with your database configuration.

## Environment Variables

The following environment variables are used for database configuration:

| Variable      | Default Value   | Description                  |
|---------------|-----------------|------------------------------|
| `DB_HOST`     | `localhost`     | Database host                |
| `DB_PORT`     | `3306`          | Database port                |
| `DB_USERNAME` | `wa_serv`       | Database username            |
| `DB_PASSWORD` | `password`      | Database password            |
| `DB_NAME`     | `db_name`       | Database name                |

## Example

Here is an example of how the system works:

1. A user sends a receipt image via WhatsApp.
2. The system processes the image, extracts receipt data, and calculates points.
3. The calculated points are stored in the database and optionally sent back to the user as a reply.

## TODO

The following features are planned for future implementation:

- Handle media image messages.
- Store images in the database.
- Implement point calculation logic.
- Integrate with an LLM (Large Language Model) to read receipt images.
- Send notifications to users after earning points.
- Handle inquiries to check earned points from users.
- Handle point redemption.

## Development Notes

This implementation was assisted by GitHub Copilot to create and refine the use case. The initial setup and implementation were completed in less than 8 working hours.

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
