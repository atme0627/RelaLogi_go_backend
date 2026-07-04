FROM golang:1.25

WORKDIR /relalogi
COPY . .

#Cを呼び出すGoパッケージの作成にcgoが必須
#デフォルトでenableのはずだが一応明示
ENV CGO_ENABLED=1

RUN apt-get update && apt-get install -y libopencv-dev libtesseract-dev pkg-config libleptonica-dev tesseract-ocr-eng

RUN go build -o app/server ./cmd
