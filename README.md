# apfs-compactor

This tool is a proof-of-concept userland tool for doing compaction of files against Apples new file-system called: APFS. *This code should be regarded as a prototype and only used on test files or data that you can easily reproduce from backups.*

This file system is applied only to those MacOS systems that have upgraded to High Sierra and have a SSD drive as their boot drive. Additionally, older HFS+ drives can be manually converted to APFS using the built-in Disk Utility tool.

## Motivation

The APFS has a few innovations around saving disk space by implementing copy-on-write, and smarter meta-data linking when it comes to file creation and copy operations.  Currently, when a drive is converted from HFS+ to APFS, there is no concept of deduplication on any existing files.  The magic happens however when files are copied. APFS, is smart enough to understand when to make logical vs physical copies of data. Additionally, it's smart enough to implement a delta copy strategy much like a Git repo does when differences are made to files.

The end result of this behavior can be a considerable disk savings and speed benefit during copy operations.

Again, no deduplication takes place by default except when attempts to copy data are made. This means, if you simply convert a drive from HFS+ to APFS you likely won't notice much of a difference initially because the conversation process will simply copy files byte for byte as seen on the source volume and as it's written to the destination volume.

That is where this tool comes: This tool will effectively compact a folder (recursively) by identifiying duplicates files first by file size (for speed purposes) then by calculating a SHA1 hash against only those files that have the same byte size.

Once these exact duplicate files are identified they'll be copied to a destination folder and upon copying the APFS, will intelligently know that these are duplicates and only copy meta-data. None of this magic happens within the Go source code as it is a feature of the APFS that is being exploited.

Once the duplicate files are copied (with the original folder hiearchy and names preserved) the non-duplicate files will additionally be copied over.

Upon completion of this operation you should have two identical folders that exist at the file-system level but due to the APFS smarts, disk space will have been intelligently used and the newly recreated destination folder will logically be the same as the source folder.

Lastly after verification the source folder can be destroyed and this is when the APFS will simply update it's Metadata database to ensure things are preserved as needed.

## Trying this out

1. Clone this repository on a USB thumb-drive that is HFS+ formatted.
2. Use Get Info on the thumb-drive and make a note of how much disk space is used.
2. Once cloned, run Disk Utility to convert the thumb-drive to APFS.
3. Run: go run main.go source_files to start the compaction process
4. Compare the source_folder and the dest_folder.
5. Finally delete the source_folder.
6. Use Get Info on the thumb-drive and observe that the drive has more free space then before the migration was applied.  The end result will actually be less overall bytes used because for those files that were copied over and identified as exact duplicates, APFS was smart enough to detect these duplicates and just perform Meta-data writes vs a real byte for byte copy.

# Notes

## Test Data
The files I was working with were generated with the following command with various settings:

```sh
# https://www.skorks.com/2010/03/how-to-quickly-generate-a-large-file-on-the-command-line-with-linux/
dd if=/dev/urandom of=file.txt bs=2048 count=10
```

These files are not anything special and opaque to this application in fact this should work on any kind of file whether it be video, audio, text, system files, etc. 

## File timestamps
I didn't bother to preserve timestamps when the files are copied over to the dest directory...so whatever happens by default with this code is the case...although I'm sure this code could be updated to better preserve "stat" based data.