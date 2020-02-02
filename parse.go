package main

/*
Declaration:

http://dsk.ippt.pan.pl/docs/abaqus/v6.13/books/usb/default.htm?startat=pt01ch03s02abx14.html#usb-int-dkeywordproc

- Continuation of a keyword line is not supported. E.g.
```
*ELASTIC, TYPE=ISOTROPIC,
DEPENDENCIES=1
```
- "Certain keywords must be used in conjunction with other keywords" is considered true, and not validated here.

- Files referenced by the INPUT or FILE parameter are not yet processed.
*/