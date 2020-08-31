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

var reqPost = function(url, data, success, failure) {
  var postData = "";
  for (var property in data) {
    if (postData !== "") postData += "&";
    postData += property  + "=" + encodeURIComponent(data[property]);
  }

  var request = new XMLHttpRequest();
  request.open("POST", url, true);
  request.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
  request.onload = function() {
    if (this.status >= 200 && this.status < 400) {
      success(JSON.parse(this.response));
    } else {
      failure();
    }
  };
  request.onerror = failure;
  request.send(postData);
};
