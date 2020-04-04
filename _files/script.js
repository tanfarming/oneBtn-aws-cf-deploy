function theOneBtnClicked() {
    input=document.getElementById("input").value
    // console.warn(input)

    // document.getElementById("console").innerHTML=""
    // refreshButton=document.createElement("BUTTON");
    // refreshButton.innerHTML="refresh"
    // refreshButton.className="btn btn-warning btn-lg";
    // refreshButton.setAttribute("onclick", "refresh()");
    // document.getElementById("console").appendChild(refreshButton);

    // refreshInfoBoard=document.createElement("div");
    // refreshInfoBoard.setAttribute("id", "info")
    // document.getElementById("console").appendChild(refreshInfoBoard);


    var xhttp = new XMLHttpRequest(); res=""
    xhttp.onreadystatechange = function() {
      if (this.readyState == 4 && this.status == 200) {
        res = this.responseText;
      }
      console.warn(res)
      document.getElementById("logbox").innerHTML += res+"&#10;"
    };
    //xhttp.open("GET", "demo_get2.asp?fname=Henry&lname=Ford", true);
    xhttp.open("POST", "oneBtn", true);
    xhttp.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    xhttp.send(input);
  }

  function refresh(){
    var xhttp = new XMLHttpRequest(); res=""
    xhttp.onreadystatechange = function() {
      if (this.readyState == 4 && this.status == 200) {
        res = this.responseText;
      } else{
        res = "errrrrrrr"
      }
      console.warn(res)
      document.getElementById("info").innerHTML = "<br>---click refresh button to refresh---<br>"+res
    };
    xhttp.open("GET", "status", true);
    // xhttp.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    xhttp.send();
  }