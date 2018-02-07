$(document).ready(function() {

    var selectedItems = [];
    var event = {}
    var socket;
    var timerStarted = false;
    //to be changed - dummy
    var event_id = "test_event";
    var nIntervId;
    init();


    $(document).on('click', '.item-box', function() {
        var id = $(this).attr('id');

        if (!$(this).hasClass('selected')) {
            processEventItem("book", id)


        } else {
            processEventItem("unbook", id)

        }
        sendMessage(JSON.stringify(event))
    })

    $(document).on('click', '.remove-item', function() {

        processEventItem("unbook", $(this).siblings('.item-id').text())
        sendMessage(JSON.stringify(event))

    })

    $(document).on('click', '#buy-places', function() {
        event.action = "buy";
        sendMessage(JSON.stringify(event))
        stopTimer()
        alert("Congrats!")
    })

    ////FUNCTIONS =====================================

    function displayItems() {
        $('.items-selected').empty();
        //populate template for a sidebar selected items
        $.each(selectedItems, function(index, value) {
            var template = document.querySelector('#selected-item-template')
            span = template.content.querySelectorAll("span")
            span[0].textContent = value;
            var is = document.querySelector(".items-selected");
            var clone = document.importNode(template.content, true);
            is.appendChild(clone)
        });
    }

    function sendMessage(message) {
        socket.send(message);
    }

    function initConnection() {
        socket = new WebSocket("ws://localhost:8000/ws");
        socket.onopen = function() {
            socket.send(event.event);
        }

        socket.onmessage = function(event) {
            var data = JSON.parse(event.data);

            if ((data.hasOwnProperty('errorCode') && data.errorCode != 0)) {
                return false;
            }

            if (data.hasOwnProperty('sysMessage') && data.messageType == 0) {
                if (!timerStarted && selectedItems.length != 0) {
                    startTimer(data.bookTime, $('#time'));
                    timerStarted = true;
                }
                return false
            }

            if (data.hasOwnProperty('sysMessage') && data.messageType == 1) {
                var position = $.inArray(data.lastActedPlace, selectedItems)
                selectedItems.splice(position, 1);
                displayItems()
                $('#' + data.lastActedPlace).removeClass('selected');
                $('#' + data.lastActedPlace).addClass('disabled-place');
                alert(data.sysMessage)
                return false;
            }

            //Clear all then re-populate
            $('.disabled-place').removeClass('disabled-place');
            // Re-populate
            $.each(data.places, function(index, value) {
                if ($.inArray(value, selectedItems) == -1) {
                    $('#' + value).addClass('disabled-place');
                }
            })
        };
    }

    function removeItem(array, element) {
        const index = array.indexOf(element);
        array.splice(index, 1);
    }

    function init() {
        event["event"] = event_id;
        event["places"] = [];
        event["action"] = "";
        initConnection();
    }

    function startTimer(duration, display) {
        var timer = duration,
            minutes, seconds;
        nIntervId = setInterval(function() {
            minutes = parseInt(timer / 60, 10);
            seconds = parseInt(timer % 60, 10);

            minutes = minutes < 10 ? "0" + minutes : minutes;
            seconds = seconds < 10 ? "0" + seconds : seconds;

            display.text(minutes + ":" + seconds);
            if (--timer < 0) {
                event["action"] = "rejected_timeout"

                sendMessage(JSON.stringify(event))

                $.each(selectedItems, function(index, value) {
                    $('#' + value).removeClass('selected');
                });

                selectedItems = [];
                displayItems()
                event["places"] = selectedItems;
                stopTimer();
                return false;
            }
        }, 1000);
    }

    function stopTimer() {
        clearInterval(nIntervId);
        $('#time').empty();
        timerStarted = false;
    }

    function processEventItem(action, id) {
        $('#' + id).toggleClass('selected');

        if (action === "book") {
            selectedItems.push(id)
        } else if (action === "unbook") {
            removeItem(selectedItems, id);

            if (selectedItems.length == 0) {
                stopTimer();
            }
        }

        displayItems();
        event["places"] = selectedItems;
        event["lastActedPlace"] = id;
        event["action"] = action;

        if ($('#userLoggedStatus').val() == 1 && selectedItems.length > 0) {
            $('#buy-places').attr('disabled', false)
        } else {
            $('#buy-places').attr('disabled', true)
        }
    }
})