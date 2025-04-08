# OpenAI Key Server

A server application that generates temporary OpenAI API keys for authorized users through Google OAuth2 authentication.

[![codecov](https://codecov.io/gh/hi120ki/openaikeyserver/branch/main/graph/badge.svg)](https://codecov.io/gh/hi120ki/openaikeyserver)

![Image](https://github.com/user-attachments/assets/20d8f5a0-7c7c-499f-b7c3-313e541826aa)

## Overview

This application provides a secure way to generate temporary OpenAI API keys for authorized users. It uses Google OAuth2 for authentication and the OpenAI Management API to create and manage API keys. Keys older than specified expiration time (default 24 hours) are automatically cleaned up.

The server provides a simple web interface for users to retrieve their keys. The application is designed to be easy to set up and use, making it ideal for personal projects or small teams.

## Features

- Google OAuth2 authentication
- OIDC verification
- Authorized user access control
- Automatic API key cleanup (keys older than specified expiration time, runs every cleanup interval, default 1 hour)
- Simple web interface for key retrieval

## Environment Variables

| Variable                | Description                                                           | Required | Default          |
| ----------------------- | --------------------------------------------------------------------- | -------- | ---------------- |
| `ALLOWED_USERS`         | Comma-separated list of email addresses allowed to access the service | No\*     | -                |
| `ALLOWED_DOMAINS`       | Comma-separated list of domains allowed to access the service         | No\*     | -                |
| `OPENAI_MANAGEMENT_KEY` | OpenAI Management API key                                             | Yes      | -                |
| `CLIENT_ID`             | Google OAuth2 client ID                                               | Yes      | -                |
| `CLIENT_SECRET`         | Google OAuth2 client secret                                           | Yes      | -                |
| `REDIRECT_URI`          | OAuth2 redirect URI                                                   | Yes      | -                |
| `DEFAULT_PROJECT_NAME`  | Default OpenAI project name                                           | No       | "personal"       |
| `PORT`                  | Server port                                                           | No       | "8080"           |
| `EXPIRATION`            | Key expiration time in seconds                                        | No       | 86400 (24 hours) |
| `CLEANUP_INTERVAL`      | Key cleanup interval in seconds                                       | No       | 3600 (1 hour)    |
| `TIMEOUT`               | HTTP client timeout in seconds                                        | No       | 10               |

\*Note: Either `ALLOWED_USERS` or `ALLOWED_DOMAINS` (or both) must be set.

## Installation

### Prerequisites

- Go 1.24 or higher
- OpenAI Management API key
- Google Cloud OAuth2 credentials

### Setup

1. Clone the repository
2. Create a `.env` file with the required environment variables
3. Run the application:

```bash
go run main.go
```

## Usage

1. Access the server at `http://localhost:8080` (or your configured port)
2. You will be redirected to Google's OAuth2 consent page
3. After authentication, if your email is in the allowed users list or your email domain is in the allowed domains list, you'll receive a temporary OpenAI API key
4. The key will be valid for the specified expiration time (default 24 hours)
5. The server will automatically clean up keys older than the expiration time (cleanup runs every hour by default)

## OpenAI Management Key Guide

The OpenAI Management Key is required to create and manage API keys programmatically. Here's how to obtain one:

1. Log in to the [OpenAI Platform](https://platform.openai.com/)
2. Navigate to the [Admin keys section](https://platform.openai.com/settings/organization/admin-keys)
3. Click on "Create new Admin key"
4. Give your key a name and click "Create admin key"
5. Copy the key (it will only be shown once) and add it to your `.env` file as `OPENAI_MANAGEMENT_KEY`

Note: Management keys have elevated permissions and should be kept secure. They can create, list, and delete API keys for your organization.

## Google Cloud OAuth2 Client Setup Guide

To set up Google OAuth2 authentication:

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to "APIs & Services" > "Credentials"
4. Click "Create Credentials" and select "OAuth client ID"
5. Select "Web application" as the application type
6. Add a name for your OAuth client
7. Add authorized redirect URIs (e.g., `http://localhost:8080/oauth2/callback`)
8. Click "Create"
9. Copy the Client ID and Client Secret
10. Add them to your `.env` file as `CLIENT_ID` and `CLIENT_SECRET`
11. Also set your redirect URI in the `.env` file as `REDIRECT_URI`

## License

MIT License
