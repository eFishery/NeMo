
# Coral File
- using YML

## How to get the whatsapp file session

Use this tools to generate the whatsapp session file [k1m0ch1/WhatsappLogin](https://github.com/k1m0ch1/WhatsappLogin)

you can download the latest binary file from [the latest release](https://github.com/k1m0ch1/WhatsappLogin/releases/)

and run with command `./WhatsappLogin -p 08123123123` and scan the QR

or

run with docker 

`docker run --rm -v sessions:/go/src/github.com/k1m0ch1/WhatsappLogin/sessions k1m0ch1/whatsapplogin:latest -p 08123123123`


# Development Environment

## How to run

`go run . 08123123123`

this will find the file `08123123123.session` in the `SESSION_DIR` environment variable by default this will goes to current directory running file

[] ~~Multi Schedule is not~~
[] ~~Multi file yml with some with no greeting or no schedule or no commands~~
[] ~~must test all the builder work well~~
[] ~~ada orang yang nyangkut gara gara status NEXT nya diem~~
[] ~~semua bentuk dokumen harus di kasih handler~~
[] ~~ada orang yg kirim gambar langsung tanpa command atau session apapun, langsung ngebreak soalnya dia mencoba coral yg tidak exist~~
[] harus dibikin cek, siapa yg kirim pesan sebelumnya, apakah bot atau user, kalau bot, cek pesannya sama ga, kalo sama jangan dikirim, ada case dia ngirim 300 pesan