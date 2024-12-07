# Secret Santa Emailer

Given a file of email recipients, this will randomize matches and email each
one their match. The application guarantees that nobody gets themselves, but
all other combinations are considered valid. Note that the program will not
indicate the matches to the operator unless explicitly told to do so.

The assignments are shuffled. Thr shuffled matches are then checked to ensure
that nobody is assigned to themselves. We just swap with the next recipient.
The whole thing is done in passes passes.

Once the assignments are established emails are send and our work is done. Back
to practicing your carols and testing out your eggnog recipe one more time, just
to be sure.

Other than the fact that we are formatting emails with an ancient and abandoned
library, this app is pretty wildly overbuild. It was able to generate a valid
match set on 100K records after 8 shuffles and approximately 2 seconds. Sending
100K emails would obviously take a while, though!

## Installation

If you have Go installed (see [golang.org](https://golang.org/dl/))

```shell
go install github.com/dangermike/secret-santa@latest
```

## Configuration

`secret-santa` gets its participants list from a JSON file containing. Each
participant shold be a member in an array with `name` and `address` parameters.
This defaults to `data/people.json`, but you can use any path you like with the
 `--source-file` parameter. Please see the [example](data/example.json) for
 details.

You must also customize your email body by creating a template file. The file
supports `{{.From}}` and `{{.To}}` which are the names pulled from the
participants JSON File. This defaults to `data/email.template`, but you can use
any path you like with the `--template-file` parameter. Please see the
[example](data/example.template) for details.

If you are using gmail (as is the default), you will need to set up an app
password. Navigate to [https://myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)
to set it up. Pass that password to the application using the `--from-password`
parameter.

## Usage

```plaintext
NAME:
   secret-santa - Secret Santa Emailer!

USAGE:
   secret-santa [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --from-name value      Display name for sender (default: "Secret Santa Robotic Elf")
   --from-address value   Email address to send from (required unless dry run)
   --from-password value  Password for the from-address (e.g. gmail password)
   --smtp-host value      Host to send to (default: "smtp.gmail.com")
   --smtp-port value      Port to send to (default: 587)
   --source-file value    JSON file containing an array of name/address objects (default: "data/people.json")
   --template-file value  Text file containing a template used for the email body (default: "data/email.template")
   --subject value        Email subject (default: "Shhhh! It's your Secret Santa assignment")
   --show-matches         Print the pairings out when sending
   --dry-run              Do not send. Implies --show-matches
   --help, -h             show help
   --version, -v          print the version

```

## Examples

### Dry run

```plaintext
$ ./secret-santa --dry-run --source-file data/example.json
INFO[0000] Shuffling...
INFO[0000] validating...
INFO[0000] match                                         from="Dee Dee" to=Tommy
INFO[0000] match                                         from=Ringo to="Dee Dee"
INFO[0000] match                                         from=George to=Paul
INFO[0000] match                                         from=Joey to=George
INFO[0000] match                                         from=Marky to=Ringo
INFO[0000] match                                         from=Paul to=Marky
INFO[0000] match                                         from=John to=Joey
INFO[0000] match                                         from=Tommy to=John
```

### Real

```plaintext
$ ./secret-santa --from-address "you@server.com" --from-password "your_pass" --source-file data/people.json
INFO[0000] Shuffling...
INFO[0000] validating...
INFO[0000] validating...
INFO[0001] sent
INFO[0002] sent
INFO[0002] sent
INFO[0003] sent
INFO[0004] sent
INFO[0005] sent
INFO[0006] sent
INFO[0007] sent
INFO[0008] sent
INFO[0009] sent
```

## Release notes

### v0.1.0 - 2024-12-06

* Configured to use go modules
* Upgraded all dependencies
* Switched from `math/rand` to `math/rand/v2`
  * Replaced custom `Shuffle` method with built-in method
* Replaced "reshuffle" with a swap with the next recipient
* Added support for `extras` in the recipients and templates

### v0.0.1 - 2019-12-12

Initial release
