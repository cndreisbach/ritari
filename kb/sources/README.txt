This directory holds raw material (transcripts, exports, PDFs) that notes
cite. Two rules keep the knowledge base a valid OKF bundle:

1. Never put .md files here. Anything ending in .md is a concept to OKF tools.
2. Never put material in hidden (dot-) directories. Other OKF tools descend
   into them and would treat a stray .md file there as a conformance error.

Cite a file from a note with its bundle path, for example:

    rit note write --title "..." --source "Call transcript /sources/call.txt"
