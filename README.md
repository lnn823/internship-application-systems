# Cloudflare Internship Application: Systems

### Junxiong Lin (linjunxiong1997@gmail.com)

## What is it?

Please write a small Ping CLI application for MacOS or Linux.
The CLI app should accept a hostname or an IP address as its argument, then send ICMP "echo requests" in a loop to the target while receiving "echo reply" messages.
It should report loss and RTT times for each sent message.

Please choose from among these languages: C/C++/Go/Rust

## Useful Links

- [A Tour of Go](https://tour.golang.org/welcome/1)
- [The Rust Programming Language](https://doc.rust-lang.org/book/index.html)

## Requirements

### 1. Use one of the specified languages

Please choose from among C/C++/Go/Rust. If you aren't familiar with these languages, you're not alone! Many engineers join Cloudflare without
specific langauge experience. Please consult [A Tour of Go](https://tour.golang.org/welcome/1) or [The Rust Programming Language](https://doc.rust-lang.org/book/index.html).

### 2. Build a tool with a CLI interface

The tool should accept as a positional terminal argument a hostname or IP address.

### 3. Send ICMP "echo requests" in an infinite loop

As long as the program is running it should continue to emit requests with a periodic delay.

### 4. Report loss and RTT times for each message

Packet loss and latency should be reported as each message received.

## Submitting your project

When submitting your project, you should prepare your code for upload to Greenhouse. The preferred method for doing this is to create a "ZIP archive" of your project folder: for more instructions on how to do this on Windows and Mac, see [this guide](https://www.sweetwater.com/sweetcare/articles/how-to-zip-and-unzip-files/).

Please provide the source code only, a compiled binary is not necessary.

## Extra Credit

1. Add support for both IPv4 and IPv6
2. Allow to set TTL as an argument and report the corresponding "time exceeded” ICMP messages
3. Any additional features listed in the ping man page or which you think would be valuable

## Junxiong Lin's GO version of ping
To run the ping program: Simply type
```bash
sudo ./ping [-6h] [-m ttl] [-c count] [-i wait] [-t timeout] [-s size] dest_ip_addr
```

Options:
```
-6 
        IPv6 (default IPv4)
-c count
    	Stop after sending (and receiving) count ECHO_RESPONSE packets. (default 2147483647)
-i wait
    	Wait wait seconds between sending each packet.  The default is to wait for one second between each packet. (default 1)
-m ttl
    	TTL limit (default 2147483647)
-s size
    	Specify the number of data bytes to be sent. (default 56)
-t timeout
    	Specify a timeout, in seconds, before ping exits regardless of how many packets have been received. (default 2147483647)
```