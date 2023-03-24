#基础镜像
FROM ubuntu:latest AS builder
LABEL stage=plat

#增加环境变量
ENV  GOARCH amd64
ENV GOOS linux
ENV  GOBIN $GOROOT/bin/
ENV GOTOOLS $GOROOT/pkg/tool/
ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn,direct
#执行命令
#RUN  sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list
RUN  apt-get clean
RUN apt-get update
RUN apt-get -y install ca-certificates
RUN sed -i "s@http://.*archive.ubuntu.com@https://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list
RUN sed -i "s@http://.*security.ubuntu.com@https://mirrors.tuna.tsinghua.edu.cn@g" /etc/apt/sources.list
RUN apt-get clean
RUN apt-get update
RUN apt-get -y install tzdata
ENV TZ Asia/Shanghai
RUN apt-get -y install vim
ARG USERNAME=wzp
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m -s /bin/bash $USERNAME \
    #
    # [Optional] Add sudo support. Omit if you don't need to install software after connecting.
    && apt-get update \
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# ********************************************************
# * Anything else you want to do like clean up goes here *
# ********************************************************

# [Optional] Set the default user. Omit if you want to keep the default as root.
USER $USERNAME
EXPOSE 8080
ADD .  /opt
# COPY ./.bashrc /home/$USERNAME/
# RUN mkdir -p /home/$USERNAME/workspace

