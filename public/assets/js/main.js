$(document).ready(function() {

    var selectedItems = [];
    var event = {}
    var socket;
    //to be changed - dummy
    var event_id = "test_event";
    init();


    $(document).on('click', '.item-box', function() {
        var id = $(this).attr('id');

        if (!$(this).hasClass('selected')) {
            $(this).addClass('selected');
            selectedItems.push(id)
            event["action"] = "book";
        } else {
            $(this).removeClass('selected');
            removeItem(selectedItems, id);
            event["action"] = "unbook";
        }

        displayItems()
        event["places"] = selectedItems;
        event["lastActedPlace"] = id;
        sendMessage(JSON.stringify(event))
    })

    $(document).on('click', '.remove-item', function() {
        var id = $(this).siblings('.item-id').text();
        removeItem(selectedItems, id);
        displayItems();
        event["places"] = selectedItems;
        event["lastActedPlace"] = id;
        event["action"] = "unbook";
        sendMessage(JSON.stringify(event))
        $('#' + id).removeClass('selected');
    })

    $(document).on('click', '#buy-places', function() {
        event.action = "buy";
        sendMessage(JSON.stringify(event))
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
            console.log(event)
            var data = JSON.parse(event.data);
            if ((data.hasOwnProperty('errorCode') && data.errorCode != 0) || (data.hasOwnProperty('sysMessage') && data.messageType == 0)) {
                return false;
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
});