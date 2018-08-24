FROM busybox:1

COPY golarn /bin/golarn
USER 1001
EXPOSE 8080
CMD /bin/golarn

