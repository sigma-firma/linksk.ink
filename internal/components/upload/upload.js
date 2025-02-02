async function submitPost() {
        const form = document.getElementById("uploadForm");
        const data = new FormData(form);
        let response = await fetch("/uploadItem", {
            method: "POST",
            body: data,
        });

        let res = await response.json();
        handleResponse(res);
    }

    function handleResponse(res) {
        if (res.success == "true") {
            window.location = window.location.origin;
        } else {
            document.getElementById("submit-butt").innerHTML = res.msg;
        }
    }
