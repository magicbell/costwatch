# Base image for Dashboard app
FROM node:22
WORKDIR /app

# Enable Corepack to use Yarn
RUN corepack enable

COPY apps/dashboard/package.json apps/dashboard/yarn.lock ./apps/dashboard/
RUN yarn --cwd ./apps/dashboard install --prefer-offline --no-progress --non-interactive --silent

# Copy the rest of the repo
WORKDIR /app
COPY . .

# No default CMD; docker-compose will run the dev server
