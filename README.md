# Twitch Chat Bot

A Golang-based application for Twitch integration and chat bot functionality.

## Overview

This project provides a robust Twitch chat bot implementation with a web interface built using Go, Templ templating, and modern frontend tooling. The application allows you to automate interactions with Twitch chat, respond to commands, and provide additional features for streamers and their communities.

## Features

- Real-time Twitch chat integration
- Custom command creation and management
- Web-based dashboard for configuration and monitoring
- Responsive UI built with Templ and Tailwind CSS
- Extensible architecture for adding custom plugins/modules

## Technology Stack

- **Backend**: Go (Golang)
- **Frontend**:
  - [Templ](https://github.com/a-h/templ) for type-safe HTML templating
  - [Tailwind CSS](https://tailwindcss.com/) for styling
  - [esbuild](https://esbuild.github.io/) for JavaScript bundling
  - Node.js for asset building
- **CI/CD**: GitHub Actions

## Prerequisites

- Go 1.18 or higher
- Node.js 22.x
- npm

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/twitch-chat-bot.git
   cd twitch-chat-bot
   ```

2. Install Go dependencies:
   ```bash
   go mod download
   ```

3. Install the Templ CLI:
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   ```

4. Install JavaScript dependencies:
   ```bash
   cd assets/js
   npm ci
   ```

## Development

### Building the application

1. Generate Templ templates:
   ```bash
   templ generate
   ```

2. Build frontend assets:
   ```bash
   cd assets/js
   npm run esbuild
   npm run tailwindcss
   ```

3. Build the Go application:
   ```bash
   go build -v ./...
   ```

### Running tests

```bash
go test -v ./...
```

### Running the application locally

```bash
go run cmd/bot/main.go
```

## Configuration

Copy the `.env.example` file to `.env` and update the example properties with your own values:

```bash
cp .env.example .env
```

Then edit the `.env` file with your specific Twitch credentials:

```
TWITCH_CLIENT_ID=your_client_id
TWITCH_CLIENT_SECRET=your_client_secret
TWITCH_BOT_USERNAME=your_bot_username
TWITCH_BOT_OAUTH_TOKEN=oauth:your_oauth_token
TWITCH_CHANNELS=channel1,channel2
```

You can obtain your Twitch credentials by creating an application in the [Twitch Developer Console](https://dev.twitch.tv/console/apps).

## Deployment

The project includes GitHub Actions workflows for CI/CD. When you push to the `master` or `develop` branches, the workflow will automatically build and test your application.

## Contributing

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin feature/my-new-feature`
5. Submit a pull request

## Acknowledgements

- [Twitch API Documentation](https://dev.twitch.tv/docs/)
- [Templ](https://github.com/a-h/templ)
- [Tailwind CSS](https://tailwindcss.com)
