FROM node:20-alpine AS deps

ARG NEXT_PUBLIC_BACKEND_URL
ENV NEXT_PUBLIC_BACKEND_URL=${NEXT_PUBLIC_BACKEND_URL}

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm install

COPY . .	

RUN NEXT_PUBLIC_BACKEND_URL=$NEXT_PUBLIC_BACKEND_URL npm run build

FROM node:20-alpine AS runner

WORKDIR /app

COPY --from=deps /app ./

EXPOSE 3000

CMD ["npm", "start"]

