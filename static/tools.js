// navbar burger menu handler
document.addEventListener('DOMContentLoaded', () => {
  const $navbarBurgers = Array.prototype.slice.call(document.querySelectorAll('.navbar-burger'), 0);
  if ($navbarBurgers.length > 0) {
    $navbarBurgers.forEach( el => {
      el.addEventListener('click', () => {
        const target = el.dataset.target;
        const $target = document.getElementById(target);
        el.classList.toggle('is-active');
        $target.classList.toggle('is-active');
      });
    });
  }
});

var encodeJSONToForm = function(data){
  var postData = "";
  for (var property in data) {
    if (postData !== "") postData += "&";
    postData += property  + "=" + encodeURIComponent(data[property]);
  }
  return postData;
};

var reqPost = function(url, data, success, failure) {
  var request = new XMLHttpRequest();
  request.open("POST", url, true);
  request.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  request.onload = function() {
    if (this.status >= 200 && this.status < 400) {
      success(this.response);
    } else {
      failure(this.response);
    }
  };
  request.onerror = function(){
    failure("Request error");
  }; 
  request.send(encodeJSONToForm(data));
};

var reqGet = function(url, success, failure) {
  var request = new XMLHttpRequest();
  request.open("GET", url, true);
  request.onload = function() {
    if (this.status >= 200 && this.status < 400) {
      success(this.response);
    } else {
      failure(this.response);
    }
  };
  request.onerror = function(){
    failure("Request error");
  }; 
  request.send();
};

var removeChildren = function(el) {
  while(el.firstChild) el.removeChild(el.firstChild);
};
