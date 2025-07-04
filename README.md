# Twitchong

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

- Go 1.24.2 or higher
- Node.js 22.x
- npm

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/OleksandrOleniuk/twitchong.git
   cd twitchong
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

### Running the application locally

```bash
air
```
### Running tests

```bash
go test -v ./...
```

## Configuration

Copy the `.env.example` file to `.env` and update the example properties with your own values:

```bash
cp .env.example .env
```

Then edit the `.env` file with your specific Twitch credentials:

```
BOT_USER_ID=bot_user_id
CHAT_CHANNEL_USER_ID=chat_channel_user_id

CLIENT_ID=cliend_id
CLIENT_SECRET=client_secret

TWITCH_SECRET_STATE=twitch_secret_state

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
