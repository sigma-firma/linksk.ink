# use in n/vim to restart on save:
# :autocmd BufWritePost * silent! !./autoload.sh
#!/bin/bash
pkill linksk.ink || true
go build -o linksk.ink
echo http://localhost:17297
./linksk.ink >> log.txt 2>&1 &
