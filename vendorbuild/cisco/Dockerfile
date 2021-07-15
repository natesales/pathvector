FROM devhub-docker.cisco.com/iox-docker/ir800/base-rootfs
RUN opkg update
RUN opkg install iox-toolchain
RUN mkdir -p /var/pathvector/
COPY pathvector_linux_amd64/pathvector /var/pathvector/pathvector
RUN chmod +x /var/pathvector/pathvector
CMD /var/pathvector/pathvector
