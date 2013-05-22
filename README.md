to run:

    btcfs > btcfs.log 2>&1
    sudo mount -t 9p 127.0.0.1 -o port=5640,version=9p2000.u,access=any /mnt

read the [kernel docs on 9p](https://www.kernel.org/doc/Documentation/filesystems/9p.txt) for more info
