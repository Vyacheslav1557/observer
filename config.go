package observer

type Config struct {
	NatsUrl   string `env:"NATS_URL" env-default:"nats://localhost:4222"`
	JwtSecret string `env:"JWT_SECRET" required:"true"`
}
