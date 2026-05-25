**SMTP Setup**

- **Goal**: Configure the `notification-service` to send emails via an external SMTP server instead of MailHog.
- **Do NOT** commit real credentials into the repository. Use a local `.env` file or environment variables.

Steps

1. Create a local `.env` from the example at the repository root:

   cp .env.example .env

2. Edit `.env` and set real credentials (example values):

   SMTP_HOST=smtp.mail.yahoo.com
   SMTP_PORT=465
   SMTP_FROM=kavix@yahoo.com
   SMTP_USER=kavix@yahoo.com
   SMTP_PASSWORD=<your-real-password>

3. Start the stack (docker-compose will pick up variables from the `.env` file):

   docker-compose up --build

4. The `notification-service` is already implemented to use `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, and `SMTP_FROM` from environment variables (see `notification-service/internal/config/config.go`). If `SMTP_HOST` is unset it defaults to `smtp.mail.yahoo.com` and `SMTP_PORT` defaults to `587`.

Notes / Security

- Keep `.env` local and never commit it. Add `.env` to `.gitignore` if not already ignored.
- For CI or production, use your platform's secret management (GitHub Actions secrets, Kubernetes secrets, etc.).
- If using port `465`, a TLS-wrapped connection is used. If using `587`, STARTTLS is attempted.
