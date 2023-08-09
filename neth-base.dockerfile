FROM ubuntu:jammy

# update repos
RUN apt-get update


# install java + teku
RUN DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get -y install tzdata
RUN apt-get install -y openjdk-17-jdk git

RUN git clone https://github.com/ConsenSys/teku.git
RUN cd teku && git checkout 23.8.0 && ./gradlew installDist
RUN cp -r teku/build/install/teku /usr/local/teku
ENV PATH="${PATH}:/usr/local/teku/bin"
RUN rm -rf teku

# install tools

RUN apt-get install -y curl jq bc wget
RUN wget https://github.com/freshautomations/stoml/releases/download/v0.7.1/stoml_linux_amd64
RUN chmod 775 stoml_linux_amd64 
RUN mv stoml_linux_amd64 /usr/local/bin/stoml

# install python + chaos-eth
RUN apt-get install -y bison build-essential cmake flex git libedit-dev \
  libllvm12 llvm-12-dev libclang-12-dev python2 zlib1g-dev libelf-dev libfl-dev python3-distutils
RUN ln /usr/bin/python2 /usr/bin/python

RUN git clone https://github.com/iovisor/bcc.git
RUN cd bcc && git checkout v0.26.0
RUN mkdir bcc/build
RUN cd bcc/build && cmake .. && make && make install && \
        cmake -DPYTHON_CMD=python .. && cd src/python/ && make && make install
RUN rm -rf bcc

RUN git clone https://github.com/javierron/royal-chaos.git
RUN cd royal-chaos && git checkout error-model-extraction