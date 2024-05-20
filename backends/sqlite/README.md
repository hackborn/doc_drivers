# SQLITE driver

Utility for generating new SQLITE drivers for [hackborn/doc](https://github.com/hackborn/doc). Use cmd/driverutil
to generate various pieces.

When developing a new driver, the workflow is:

- Make changes to the reference driver in ref/\*
- Run driverutility->Run reference driver to see if it worked.
- Run driverutility->Make templates to make new templates used to generate new drivers. This will read the ref/_ files and convert them into templates/_ files.
- Compile
- Run driverutility->Make driver to make a new testable driver from the templates. This will use the templates/_ files to create the gen/_ files.
- Compile
- Run driverutility->Test driver to verify the generated driver works.
