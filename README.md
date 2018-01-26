# Ticket booking project

Simple project (skeleton) of a ticket booking system.
Consists only of one event (accessible via http://localhost/event) and predefined several places to choose from.
User can book or buy(after booking) tickets for places.

## Requirements
- User possible interactions (like observing in his browser the places that is booked by other users) is performed using Websockets.
- User can book a place(seat) or several places. These selected places will be disabled if other users is currently also selecting places on the same event.
- If it somehow happened that users simultaniously clicked on the same place - system will process the first request received by system and will give other user response with answer that this place is already booked by someone else.
