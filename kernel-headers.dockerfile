ARG target
FROM javierron/neth:${target}

RUN apt-get install -y linux-headers-$(uname -r)