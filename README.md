# Crystal Sage

<p align="center">
  <img src="https://github.com/olivatooo/crystal-sage/blob/main/logo.png?raw=true" alt="logo"/>
</p>

Configure all of your external logs in only one place. Tired of trying to understand how slack, discord, and telegram work? Tired of creating a new webhook and a new application everytime you want to fire logs? This is for you!

## One yaml to rule them all

Configure your logs in one place.

## How it works

Configure YAML -> Make request to Crystal Sage ->  [Slack, Discord, Telegram, Your Mom]

---

## Quick Start

### Using Docker (Recommended)

The easiest way to run Crystal Sage is using the pre-built Docker image:

```bash
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  olivatooo/crystal-sage:latest
```

Or with environment variables:

```bash
docker run -d \
  -p 8080:8080 \
  -e YAML_PATH=/app/config.yaml \
  -v $(pwd)/config.yaml:/app/config.yaml \
  olivatooo/crystal-sage:latest
```

### Building from Source

1. **Prerequisites**: Go 1.23 or later

2. **Clone the repository**:
   ```bash
   git clone https://github.com/olivatooo/crystal-sage.git
   cd crystal-sage
   ```

3. **Install dependencies**:
   ```bash
   go mod download
   ```

4. **Build the binary**:
   ```bash
   go build -o crystal-sage .
   ```

5. **Run**:
   ```bash
   ./crystal-sage
   ```

   By default, it reads `config.yaml` from the current directory. You can override this with the `YAML_PATH` environment variable:
   ```bash
   YAML_PATH=/path/to/config.yaml ./crystal-sage
   ```

---

## Getting Webhook Credentials

### Telegram Bot Token and Chat ID

1. **Create a Telegram Bot**:
   - Open Telegram and search for [@BotFather](https://t.me/botfather)
   - Send `/newbot` command
   - Follow the instructions to name your bot
   - BotFather will provide you with a **bot token** (format: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

2. **Get your Chat ID**:
   - Start a conversation with your bot
   - Send any message to your bot
   - Visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
   - Look for `"chat":{"id":123456789}` in the response - that's your **chat ID**

3. **Configure in Crystal Sage**:
   ```yaml
   webhook: "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/sendMessage"
   chatId: "<YOUR_CHAT_ID>"
   ```

### Discord Webhook

1. **Create a Discord Webhook**:
   - Open your Discord server
   - Go to **Server Settings** → **Integrations** → **Webhooks**
   - Click **New Webhook**
   - Configure the webhook (name, channel, avatar)
   - Click **Copy Webhook URL**

2. **Configure in Crystal Sage**:
   ```yaml
   webhook: "https://discord.com/api/webhooks/123456789/abcdefghijklmnopqrstuvwxyz"
   ```

### Slack Webhook

1. **Create a Slack Incoming Webhook**:
   - Go to [api.slack.com/apps](https://api.slack.com/apps)
   - Click **Create New App** → **From scratch**
   - Name your app and select your workspace
   - Go to **Incoming Webhooks** → Enable it
   - Click **Add New Webhook to Workspace**
   - Select the channel and click **Allow**
   - Copy the **Webhook URL**

2. **Configure in Crystal Sage**:
   ```yaml
   webhook: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
   ```

---

## Configuration

Create a `config.yaml` file:

```yaml
global:
  debug: true  # Whether to print debug logs in terminal
  port: 8080   # Port to listen on

crystals:
  - name: "my-crystal"
    shards:
      # Discord example
      - alias: "Discord Notifications"
        type: "discord"
        envVar: false
        webhook: "https://discord.com/api/webhooks/..."

      # Telegram example
      - alias: "Telegram Alerts"
        type: "telegram"
        envVar: false
        webhook: "https://api.telegram.org/bot<TOKEN>/sendMessage"
        chatId: "123456789"

      # Slack example
      - alias: "Slack Logs"
        type: "slack"
        envVar: true
        webhook: "SLACK_WEBHOOK_ENV_VAR"
```

### Using Environment Variables

Set `envVar: true` to load webhook URLs from environment variables:

```yaml
- alias: "Secure Webhook"
  type: "discord"
  envVar: true
  webhook: "DISCORD_WEBHOOK_URL"  # This will read from $DISCORD_WEBHOOK_URL
```

For Telegram, you can also use environment variables for `chatId`:

```yaml
- alias: "Telegram Bot"
  type: "telegram"
  envVar: true
  webhook: "TELEGRAM_BOT_URL"
  chatId: "TELEGRAM_CHAT_ID"  # Will read from $TELEGRAM_CHAT_ID if envVar is true
```

---

## Kubernetes Deployment

### Deploying Crystal Sage on Kubernetes

Crystal Sage can be easily deployed on Kubernetes using ConfigMaps or Secrets to inject the configuration file.

#### Option 1: Using ConfigMap (Recommended for non-sensitive configs)

1. **Create a ConfigMap from your config.yaml**:
   ```bash
   kubectl create configmap crystal-sage-config \
     --from-file=config.yaml=./config.yaml \
     -n your-namespace
   ```

2. **Deploy using the provided manifests**:

   **deployment.yaml**:
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: crystal-sage
     namespace: your-namespace
   spec:
     replicas: 1
     selector:
       matchLabels:
         app: crystal-sage
     template:
       metadata:
         labels:
           app: crystal-sage
       spec:
         containers:
         - name: crystal-sage
           image: olivatooo/crystal-sage:latest
           imagePullPolicy: Always
           ports:
           - containerPort: 8080
             name: http
           volumeMounts:
           - name: config
             mountPath: /app/config.yaml
             subPath: config.yaml
             readOnly: true
           env:
           - name: YAML_PATH
             value: /app/config.yaml
           resources:
             requests:
               memory: "64Mi"
               cpu: "100m"
             limits:
               memory: "128Mi"
               cpu: "200m"
         volumes:
         - name: config
           configMap:
             name: crystal-sage-config
   ```

   **service.yaml**:
   ```yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: crystal-sage
     namespace: your-namespace
   spec:
     selector:
       app: crystal-sage
     ports:
     - port: 8080
       targetPort: 8080
       protocol: TCP
     type: ClusterIP  # Use LoadBalancer or NodePort for external access
   ```

3. **Apply the manifests**:
   ```bash
   kubectl apply -f deployment.yaml
   kubectl apply -f service.yaml
   ```

#### Option 2: Using Secret (For sensitive configurations)

If your config contains sensitive information, use a Secret instead:

```bash
# Create secret from config file
kubectl create secret generic crystal-sage-config \
  --from-file=config.yaml=./config.yaml \
  -n your-namespace
```

Then update the deployment to use a Secret volume:

```yaml
volumes:
- name: config
  secret:
    secretName: crystal-sage-config
```

#### Option 3: Using Environment Variables for Webhooks

For maximum security, store webhook URLs in Kubernetes Secrets and reference them in your config:

1. **Create secrets for webhook URLs**:
   ```bash
   kubectl create secret generic webhook-secrets \
     --from-literal=discord-webhook='https://discord.com/api/webhooks/...' \
     --from-literal=slack-webhook='https://hooks.slack.com/services/...' \
     --from-literal=telegram-bot-url='https://api.telegram.org/bot.../sendMessage' \
     --from-literal=telegram-chat-id='123456789' \
     -n your-namespace
   ```

2. **Update your config.yaml to use environment variables**:
   ```yaml
   crystals:
     - name: "my-crystal"
       shards:
         - alias: "Discord Notifications"
           type: "discord"
           envVar: true
           webhook: "DISCORD_WEBHOOK"
         - alias: "Slack Logs"
           type: "slack"
           envVar: true
           webhook: "SLACK_WEBHOOK"
         - alias: "Telegram Alerts"
           type: "telegram"
           envVar: true
           webhook: "TELEGRAM_BOT_URL"
           chatId: "TELEGRAM_CHAT_ID"
   ```

3. **Update deployment to inject secrets as environment variables**:
   ```yaml
   env:
   - name: YAML_PATH
     value: /app/config.yaml
   - name: DISCORD_WEBHOOK
     valueFrom:
       secretKeyRef:
         name: webhook-secrets
         key: discord-webhook
   - name: SLACK_WEBHOOK
     valueFrom:
       secretKeyRef:
         name: webhook-secrets
         key: slack-webhook
   - name: TELEGRAM_BOT_URL
     valueFrom:
       secretKeyRef:
         name: webhook-secrets
         key: telegram-bot-url
   - name: TELEGRAM_CHAT_ID
     valueFrom:
       secretKeyRef:
         name: webhook-secrets
         key: telegram-chat-id
   ```

### Accessing Crystal Sage from Kubernetes Services

Once deployed, other services in your Kubernetes cluster can call Crystal Sage using the service name:

**Service Discovery**:
- **Service Name**: `crystal-sage`
- **Namespace**: `your-namespace` (or `default` if not specified)
- **Port**: `8080`
- **Full DNS**: `crystal-sage.your-namespace.svc.cluster.local:8080`

**Example API Calls from Pods**:

```bash
# From within a pod in the same namespace
curl -X POST http://crystal-sage:8080/my-crystal \
  -d "content=Hello from Kubernetes!"

# From a pod in a different namespace
curl -X POST http://crystal-sage.your-namespace.svc.cluster.local:8080/my-crystal \
  -d "content=Hello from another namespace!"

# Using GET request
curl "http://crystal-sage:8080/my-crystal?content=Hello from Kubernetes!"
```

**Example in a Deployment** (calling from another service):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    spec:
      containers:
      - name: my-app
        image: my-app:latest
        command:
        - /bin/sh
        - -c
        - |
          # Send log to Crystal Sage
          curl -X POST http://crystal-sage:8080/my-crystal \
            -d "content=Application started successfully"
```

**Example using environment variables**:

```yaml
env:
- name: CRYSTAL_SAGE_URL
  value: "http://crystal-sage:8080"
- name: CRYSTAL_NAME
  value: "my-crystal"
```

Then in your application code:
```bash
curl -X POST ${CRYSTAL_SAGE_URL}/${CRYSTAL_NAME} \
  -d "content=Log message here"
```

### API Endpoints

Crystal Sage exposes HTTP endpoints based on your configuration:

- **Endpoint Pattern**: `/{crystal-name}`
  - Where `{crystal-name}` is the `name` field from your `config.yaml` crystals section
- **Methods**: `GET` and `POST`
- **Parameters**:
  - `content` (required): The message to send to all configured shards

**Example**:
If your config.yaml has:
```yaml
crystals:
  - name: "production-logs"
  - name: "alerts"
```

Then your services can call:
- `http://crystal-sage:8080/production-logs?content=Log message`
- `http://crystal-sage:8080/alerts?content=Alert message`

### Updating Configuration

To update the configuration:

1. **Update ConfigMap**:
   ```bash
   kubectl create configmap crystal-sage-config \
     --from-file=config.yaml=./config.yaml \
     -n your-namespace \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

2. **Restart pods to pick up changes**:
   ```bash
   kubectl rollout restart deployment/crystal-sage -n your-namespace
   ```

Or use a rolling update:
```bash
kubectl set env deployment/crystal-sage CONFIG_UPDATED=$(date +%s) -n your-namespace
```

---

## Usage

After starting Crystal Sage, send logs via HTTP POST or GET:

```bash
# POST request
curl -X POST http://localhost:8080/my-crystal \
  -d "content=Hello from Crystal Sage!"

# GET request
curl "http://localhost:8080/my-crystal?content=Hello from Crystal Sage!"
```

The message will be sent to all configured shards (Discord, Slack, Telegram) for that crystal.

### Calling from Kubernetes Services

From within your Kubernetes cluster, use the service DNS name:

```bash
# Same namespace
curl -X POST http://crystal-sage:8080/my-crystal \
  -d "content=Hello from Kubernetes!"

# Different namespace
curl -X POST http://crystal-sage.your-namespace.svc.cluster.local:8080/my-crystal \
  -d "content=Hello from another namespace!"
```

---

## Development

### Setting up Development Environment

1. **Fork and clone**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/crystal-sage.git
   cd crystal-sage
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Create a test config**:
   - Copy `config.yaml` and modify it with your test webhooks
   - Use `envVar: true` with environment variables for sensitive credentials

4. **Run locally**:
   ```bash
   go run main.go
   ```

### Adding a New Shard

1. Create a new file in `internal/shards/` (e.g., `custom.go`)
2. Implement the shard struct with `*Shard` embedded:
   ```go
   type Custom struct {
       *Shard
   }
   ```
3. Implement `Log(content string, level uint8)` and `RawLog(content string)` methods
4. Add a case in `internal/orb.go` switch statement to handle your shard type

### Running Tests

```bash
go test ./...
```

### Code Style

- Follow Go standard formatting: `gofmt -w .`
- Use `golint` or `golangci-lint` for linting
- Keep functions focused and testable

---

## Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the repository**
2. **Create a feature branch**:
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes**:
   - Write clean, documented code
   - Add tests if applicable
   - Update README if needed
4. **Commit your changes**:
   ```bash
   git commit -m "Add amazing feature"
   ```
5. **Push to your fork**:
   ```bash
   git push origin feature/amazing-feature
   ```
6. **Open a Pull Request**

### Contribution Guidelines

- Keep PRs focused on a single feature or fix
- Update documentation for new features
- Ensure code compiles and passes tests
- Follow existing code style and patterns
- Add examples in PR description when applicable

---

## Support

For issues, questions, or contributions, please open an issue on [GitHub](https://github.com/olivatooo/crystal-sage/issues).
