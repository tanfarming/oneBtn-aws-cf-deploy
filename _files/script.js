
//============================== =====inits=================================
document.getElementById("theOneBtn").addEventListener("click", theOneBtnClicked);

window.addEventListener("load", streamLogs);

//====================================funcs=================================
function theOneBtnClicked() {
    input=document.getElementById("input").value
    var xhttp = new XMLHttpRequest(); res=""
    xhttp.onreadystatechange = function() {
      if (this.readyState == 4 && this.status == 200) {
        res = this.responseText;
      }
    };
    xhttp.open("POST", "/OneBtnDep", true);
    xhttp.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    xhttp.send(input);
  }

function streamLogs(){
  var source = new EventSource("/LogStream");
  divLogBoard=document.getElementById("divLogBoard");
  source.onmessage = function (event) {
    // console.warn(event.data)
    divLogBoard.innerHTML+=event.data +"<br>";
    // logBoard.scrollTop = logBoard.scrollHeight;
  }
}
