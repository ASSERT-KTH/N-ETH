FROM javierron:neth-base

# install go + geth
RUN apt-get install -y gcc make

RUN wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
ENV PATH="${PATH}:/usr/local/go/bin"
RUN rm go1.19.3.linux-amd64.tar.gz

RUN git clone https://github.com/ethereum/go-ethereum.git
RUN cd go-ethereum && git checkout v1.12.0 && make geth
RUN cp go-ethereum/build/bin/geth /usr/local/bin/geth
RUN rm -rf go-ethereum

COPY ./*.sh /
COPY ./config.toml /
COPY ./*.go /
COPY ./*.json /
# CMD bash single-version-controller.sh