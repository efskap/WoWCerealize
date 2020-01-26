# Cerealize

Streams data out of a WoW addon into the outside world via a thin strip of pixels at the bottom right corner of the screen.

Currently it just subscribes to all events in the API and sends em over.

## Why?

WoW addons normally can't communicate with external programs, besides SavedVariables that are written to disk upon logout, and for good reason. But I was curious if I could implement a protocol to send data outside strictly visually.

I don't believe I'm doing botting scum any favours, since they inject their garbage into the process anyway, (or call ReadProcessMemory?). Since this is a one way channel, controlling the character with emulated input is a whole other bag of worms anyway.

### Application Ideas

-  Change the colour of your fancy RGB mouse/keyboard based on your mana/hp values, druid form, warrior stance, etc.

-  Display your guy's position on a world map on another screen.

## Details

Messages are sent with `Cerealize_Send("Hello world!")`

Not doing any fancy bitfield stuff. Messages that don't fit into the pixel strip get chunked up into several packets. `\n` indicates the end of a message (like serial I think).

Byte 1: Packet number

Byte 2: Checksum (Sum of all bytes except the first two, modulo 0xFF)

Rest of bytes: String encoded as UTF-8