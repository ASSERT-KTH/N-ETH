FROM ubuntu:focal

# update repos
RUN apt-get update


# install java + teku
RUN DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get -y install tzdata
RUN apt-get install -y openjdk-11-jdk git

RUN git clone https://github.com/ConsenSys/teku.git
RUN cd teku && git checkout 22.10.1 && ./gradlew installDist
RUN cp -r teku/build/install/teku /usr/local/teku
ENV PATH="${PATH}:/usr/local/teku/bin"
RUN rm -rf teku

# install tools

RUN apt-get install -y curl jq bc wget
RUN wget https://github.com/freshautomations/stoml/releases/download/v0.7.1/stoml_linux_amd64
RUN chmod 775 stoml_linux_amd64 
RUN mv stoml_linux_amd64 /usr/local/bin/stoml

# install python + chaos-eth
RUN apt-get install -y python
RUN apt-get install -y bison build-essential cmake flex git libedit-dev \
  libllvm12 llvm-12-dev libclang-12-dev python zlib1g-dev libelf-dev libfl-dev python3-distutils

RUN git clone https://github.com/iovisor/bcc.git
RUN mkdir bcc/build
RUN cd bcc/build && cmake .. && make && make install && \
        cmake -DPYTHON_CMD=python .. && cd src/python/ && make && make install
RUN rm -rf bcc

RUN git clone https://github.com/javierron/royal-chaos.git
RUN cd royal-chaos && git checkout error-model-extraction

# install dotnet + nethermind

RUN wget https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
RUN dpkg -i packages-microsoft-prod.deb
RUN packages-microsoft-prod.deb

RUN apt-get update 
RUN apt-get install -y dotnet-sdk-6.0 libsnappy-dev libc6-dev libc6 librocksdb5.17

RUN git clone https://github.com/nethermindeth/nethermind --recursive
RUN cd nethermind && git checkout 1.14.5

# install go for scripts
RUN wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
ENV PATH="${PATH}:/usr/local/go/bin"
RUN rm go1.19.3.linux-amd64.tar.gz

COPY ./*.sh /
COPY ./config.toml /
COPY ./*.go /
# CMD bash single-version-controller.sh