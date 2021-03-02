FROM alpine
ADD gendata /gendata
ENTRYPOINT [ "/gendata" ]
