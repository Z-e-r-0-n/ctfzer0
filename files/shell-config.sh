#global shell config
alias ls='ls --color=auto'
alias ll='ls -lah --color=auto'
alias grep='grep --color=auto'
alias cls='clear'
alias poweroff='shutdown now'
alias reboot='reboot'
# just colour diff between root and non root
if [ "$(id -u)" -eq 0 ]; then
    PS1='\[\e[31m\][ROOT@\h \W]#\[\e[0m\] '
else
    PS1='\[\e[32m\][\u@\h \W]$\[\e[0m\] '
fi
export YOUROS=1

