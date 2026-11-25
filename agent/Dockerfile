FROM node:22-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY server.ts ./
COPY prompts ./prompts
USER node
CMD ["npm","start"]
