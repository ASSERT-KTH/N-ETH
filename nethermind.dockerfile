FROM javierron:neth-base

# install dotnet + nethermind

RUN wget https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
RUN dpkg -i packages-microsoft-prod.deb
RUN rm packages-microsoft-prod.deb

RUN apt-get update 
RUN apt-get install -y dotnet-sdk-6.0 libsnappy-dev libc6-dev libc6 librocksdb5.17

RUN git clone https://github.com/nethermindeth/nethermind --recursive
RUN cd nethermind && git checkout 1.20.1
RUN cd nethermind/src/Nethermind/Nethermind.Runner && dotnet build -c Release
RUN mkdir /usr/local/nethermind
RUN cp -r nethermind/src/Nethermind/Nethermind.Runner/bin /usr/local/nethermind/bin
RUN cp -r nethermind/src/Nethermind/Nethermind.Runner/configs /usr/local/nethermind/configs
ENV PATH="${PATH}:/usr/local/nethermind/bin/Release/net6.0"

# install go for scriptss
RUN wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
ENV PATH="${PATH}:/usr/local/go/bin"
RUN rm go1.19.3.linux-amd64.tar.gz

COPY ./*.sh /wrkspc/
COPY ./config.toml /wrkspc/
COPY ./*.go /wrkspc/
COPY ./*.json /wrkspc/
# CMD bash single-version-controller.sh
