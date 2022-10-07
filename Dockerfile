FROM golang:1.19
WORKDIR cumin
COPY . .
RUN go install -buildvcs=false
