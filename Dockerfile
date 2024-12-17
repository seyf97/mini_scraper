FROM alpine:latest
WORKDIR /tmp/app

COPY build/app-amd64-linux .
COPY test_in.csv .

CMD ["./app-amd64-linux", "-i", "test_in.csv", "-o", "links_out.csv" ]