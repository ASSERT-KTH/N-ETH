FROM javierron:/neth:base

WORKDIR /wrkspc

# install dotnet + nethermind

RUN apt-get update 
RUN apt-get install -y dotnet-sdk-7.0 libsnappy-dev libc6-dev libc6

RUN git clone https://github.com/nethermindeth/nethermind --recursive
RUN cd nethermind && git checkout 1.20.1
RUN cd nethermind/src/Nethermind/Nethermind.Runner && dotnet build -c Release
RUN mkdir /usr/local/nethermind
RUN cp -r nethermind/src/Nethermind/Nethermind.Runner/bin /usr/local/nethermind/bin
RUN cp -r nethermind/src/Nethermind/Nethermind.Runner/configs /usr/local/nethermind/configs
ENV PATH="${PATH}:/usr/local/nethermind/bin/Release/net7.0"

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
