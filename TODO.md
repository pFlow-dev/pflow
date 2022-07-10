WIP
---

building a good dev tool + html doc generator

Static hosted Models
--------------------

- [ ] nest markdown in models.json
- [ ] show source and markdown on view page
- [ ] warn when duplicate schemas are found in the same collection

backlog
-------
- [ ] test to see if nested site must be found on host vs packed in binary
- [ ] can we get rid of double loading - think this is a race during boot
- [ ] do we need to show errors on front-end?

DONE
----
- [x] refactor to use paths for /image/XXXXXXX.svg
- [x] support multiple models in a collection export to html
- [x] fix bug - why does counter.js not show up? (schema conflict?)
