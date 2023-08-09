FROM javierron:neth-base

# install go + erigon
RUN apt-get install -y gcc make

RUN wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
ENV PATH="${PATH}:/usr/local/go/bin"
RUN rm go1.19.3.linux-amd64.tar.gz

RUN git clone https://github.com/ledgerwatch/erigon.git
RUN cd erigon && git checkout v2.48.1 && make erigon
RUN cp erigon/build/bin/erigon /usr/local/bin/erigon
RUN rm -rf erigon

COPY ./*.sh /
COPY ./config.toml /
COPY ./*.go /
COPY ./*.json /
# CMD bash single-version-controller.sh