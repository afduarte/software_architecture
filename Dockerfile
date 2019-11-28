# this is our first build stage, it will not persist in the final image
FROM golang as build

WORKDIR /

RUN go get -u "github.com/gin-gonic/gin"

COPY *.go ./

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main  *.go

FROM scratch

WORKDIR /

COPY --from=build /main /main

EXPOSE 80

CMD ["/main"]
