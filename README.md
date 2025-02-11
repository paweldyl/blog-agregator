Project Setup Guide

This document provides step-by-step instructions to set up PostgreSQL, install necessary dependencies, and run database migrations for this project.

Prerequisites

Ensure you have the following installed:

Ubuntu/Debian-based system (or equivalent Linux environment)

Go (installed and configured)

PostgreSQL

1. Install PostgreSQL

Run the following commands to install PostgreSQL and its contrib package:

sudo apt update
sudo apt install postgresql postgresql-contrib

2. Set PostgreSQL Password

Set a password for the postgres user:

sudo passwd postgres

3. Start PostgreSQL Server

Start the PostgreSQL service:

sudo service postgresql start

4. Access PostgreSQL Shell

Switch to the postgres user and enter the shell:

sudo -u postgres psql -d gator

5. Create a Database

Run the following command inside the PostgreSQL shell to create a new database:

CREATE DATABASE gator;

6. Install Go Dependencies

Ensure you have Go installed, then install the necessary packages:

go get github.com/google/uuid
go get github.com/lib/pq
go get github.com/joho/godotenv

7. Run Database Migrations

Use goose to run database migrations. Replace <connection_string> with your actual connection string:

goose postgres "postgres://username:password@localhost:5432/gator" up

8. Generate SQLC Go Package

Run the following command to generate the SQLC Go package:

sqlc generate

9. Commands Overview

The application provides a command-line interface to interact with users, feeds, and posts. Below are the available commands:

login [username] - Logs in as the specified user.

register [username] - Registers a new user.

reset - Resets all user data.

users - Lists all registered users.

agg [interval] - Fetches feeds at the specified interval.

addfeed [name] [url] - Adds a new feed.

feeds - Lists all available feeds.

follow [feed_url] - Follows a specific feed.

following - Lists feeds the user is following.

unfollow [feed_url] - Unfollows a specific feed.

browse - Displays posts from followed feeds.

Notes

Ensure your PostgreSQL service is running before running migrations.

Update your .env file with the correct database connection details.

Install goose if not already installed (go install github.com/pressly/goose/v3/cmd/goose@latest).

Troubleshooting

If PostgreSQL fails to start, check the logs:

sudo systemctl status postgresql

If psql command is not found, ensure PostgreSQL is installed and added to your PATH.

