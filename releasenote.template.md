## Installation

### Linux

For RHEL/CentOS:


```bash
# x86_64
sudo yum install https://github.com/masahide/assumer/releases/latest/download/__amd64rpm__

# ARM
sudo yum install https://github.com/masahide/assumer/releases/latest/download/__arm64rpm__
```


For Ubuntu/Debian:

```bash
# x86_64
wget -qO /tmp/assumer.deb https://github.com/masahide/assumer/releases/latest/download/__amd64deb__
sudo dpkg -i /tmp/assumer.deb

# ARM
wget -qO /tmp/assumer.deb https://github.com/masahide/assumer/releases/latest/download/__arm64deb__
sudo dpkg -i /tmp/assumer.deb
```

### MacOS


```bash
# x86_64
brew install masahide/tap/assumer-x86_64

# ARM (Apple silicon)
brew install masahide/tap/assumer-arm64
```
