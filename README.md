# emime

emime is an email / [MIME](https://tools.ietf.org/html/rfc2045) parser written in `Go`.

It is simplified from [enmime](https://github.com/jhillyerd/enmime), and improves to provide better support for parsing `message/rfc822`. 

The parsed tree structure (ignoring minor string differences) is the same as Python's [email parser](https://docs.python.org/3/library/email.parser.html).
