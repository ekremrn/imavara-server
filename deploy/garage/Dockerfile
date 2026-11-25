FROM dxflrs/garage:v2.5.0 AS garage
FROM alpine:3.22
COPY --from=garage /garage /garage
COPY init.sh /init.sh
CMD ["/garage","server"]
