var initTagsInput = function(tagsInput, tagsDropdown, tagsList, suggestTags) {

  var addTag = function(tag) {
    tag = tag.trim();
    if (tag === "") return;
    var removeTag = $('<button type="button" class="close" aria-label="Close"><span aria-hidden="true">&times;</span></button>');
    tagsList.append($('<span class="badge badge-light tag"></span>').append($('<span></span>').text(tag)).append(removeTag));
  }

  var getTags = function() {
    return tagsList.find('span.badge>span')
      .map(function() { return $( this ).text(); })
      .get();
  }

  var updateDropdownMenu = function(focused) {
    if (focused === undefined)
      focused = tagsInput.is(":focus");
    tagsDropdown.empty();
    var userInput = tagsInput.val().toLowerCase();
    var autoCompleteSuggestions = suggestTags(userInput);
    if (autoCompleteSuggestions.length == 0 || !focused) {
      tagsDropdown.removeClass('show');
    } else {
      tagsDropdown.addClass('show');
    }
    autoCompleteSuggestions.forEach(function (tag) {
      tagsDropdown.append($('<a class="dropdown-item" href="#" tabindex="0"></a>').text(tag));
    });
  }

  tagsInput.keypress(function(e){
    if (e.key === "," || e.key === "Enter"){
      e.preventDefault();
      var tags = tagsInput.val().split(",");
      var tag = tags.shift();
      tagsInput.val(tags.join(","));
      addTag(tag);
    }
  });
  tagsInput.keyup(function(e){
    updateDropdownMenu();
  });
  tagsInput.focus(function(e){
    updateDropdownMenu(true);
  })
  tagsInput.blur(function(e){
    if (tagsDropdown.children().is(e.relatedTarget))
      return;
    tagsDropdown.removeClass('show');
  })
  $(document).on('click', 'a.dropdown-item', function(event) {
    event.preventDefault();
    var autocompleteItem = $(this);
    addTag(autocompleteItem.text());
    tagsInput.val("");
    updateDropdownMenu();
  });

  $(document).on('click', '.tag>.close', function(event) {
    event.preventDefault();
    var close = $(this);
    close.parent().remove();
  });

  updateDropdownMenu();
  return {addTag: addTag, getTags: getTags};
}
