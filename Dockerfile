FROM golang

RUN go get github.com/dchest/captcha 
RUN go get -u github.com/go-sql-driver/mysql

EXPOSE 8080

CMD ["run"]
