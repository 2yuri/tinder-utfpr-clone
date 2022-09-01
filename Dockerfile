FROM golang:buster as builder

WORKDIR /app/tinder

COPY . ./

RUN echo "Version running at: $VERSION" 
RUN go mod tidy
CMD go build main.go && ./main