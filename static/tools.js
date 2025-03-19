function updateValueFromJson(json) {
    var response = JSON.parse(json)
    for (var prop in response) {
        var ele = document.getElementById(prop)
        if (ele) {
            if (ele instanceof HTMLInputElement) {
                ele.value = response[prop]
            } else {
                ele.textContent = response[prop]
            }
            console.log(`${prop} set to '${response[prop]}'`)
        }
        else {
            console.log(`${prop} not found!`)
        }
    }
}

function getData(url, callback = null) {
    var xhr = new XMLHttpRequest()
    xhr.open('GET', url, true)

    xhr.onreadystatechange = function () {
        if (this.readyState === XMLHttpRequest.DONE) {
            if (callback != null) {
                callback(this.status, xhr.responseText)
            } else if (this.status === 200) {
                updateValueFromJson(xhr.responseText)
            }
        }
    }

    xhr.send()

}

function postData(url, formData, callback = null) {
    console.log("postData " + formData)

    var xhr = new XMLHttpRequest()
    xhr.open('POST', url, true)
    xhr.setRequestHeader("Content-Type", "application/json")

    xhr.onreadystatechange = function () {
        if (this.readyState === XMLHttpRequest.DONE) {
            if (callback != null) {
                callback(this.status, xhr.responseText)
            }
            else if (this.status != 200)
                alert("POST failed!")
        }
    }

    xhr.send(formData)
}

function putData(url, formData) {
    console.log("putData " + formData)

    var xhr = new XMLHttpRequest()
    xhr.open('PUT', url, true)
    xhr.setRequestHeader("Content-Type", "application/json")

    xhr.onreadystatechange = function () {
        if (this.readyState === XMLHttpRequest.DONE) {
            if (this.status != 200)
                alert("PUT failed!")
        }
    }

    xhr.send(formData)
}

function deleteData(url, formData, callback = null) {
    console.log("deleteData " + formData)

    var xhr = new XMLHttpRequest()
    xhr.open('DELETE', url, true)

    if (formData != null) {
        xhr.setRequestHeader("Content-Type", "application/json")
    }

    xhr.onreadystatechange = function () {
        if (this.readyState === XMLHttpRequest.DONE) {
            if (callback != null) {
                callback(this.status, xhr.responseText)
            }
            else if (this.status != 200)
                alert("DELETE failed!")
        }
    }

    xhr.send(formData)
}
