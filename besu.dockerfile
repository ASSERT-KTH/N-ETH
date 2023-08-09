FROM javierron:neth-base

# install besu
RUN git clone https://github.com/hyperledger/besu.git
RUN cd besu && git checkout 23.7.0 && ./gradlew installDist
RUN cp -r besu/build/install/besu /usr/local/besu
ENV PATH="${PATH}:/usr/local/besu/bin"
RUN rm -rf besu

# install go for scripts
RUN wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
ENV PATH="${PATH}:/usr/local/go/bin"
RUN rm go1.19.3.linux-amd64.tar.gz

COPY ./*.sh /
COPY ./config.toml /
COPY ./*.go /
COPY ./*.json /
# CMD bash single-version-controller.sh