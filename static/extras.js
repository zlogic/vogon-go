var initTagsInput = function(tagsInput, tagsDropdown, tagsList, suggestTags) {
  var focused = false;

  var addTag = function(tag) {
    tag = tag.trim();
    if (tag === "") return;
    var tagElement = document.createElement("span");
    tagElement.setAttribute("class", "tag");
    tagElement.textContent = tag;
    var deleteTagButton = document.createElement("button");
    deleteTagButton.setAttribute("class", "delete is-small");
    deleteTagButton.type = "button";

    deleteTagButton.addEventListener("click", function(event){
      event.preventDefault();
      tagElement.remove();
    });

    tagElement.append(deleteTagButton);
    tagsList.append(tagElement);
    tagsList.append(" ")
  };

  var getTags = function() {
    var tagList = [];
    tagsList.querySelectorAll('span.tag').forEach(function(tagElement) { tagList.push(tagElement.textContent); });
    return tagList;
  };

  var updateDropdownMenu = function() {
    var suggestionsList = tagsDropdown.querySelector(".dropdown-content");
    removeChildren(suggestionsList);
    var userInput = tagsInput.value.toLowerCase();
    var autoCompleteSuggestions = suggestTags(userInput);
    if (autoCompleteSuggestions.length == 0 || !focused || userInput === "") {
      tagsDropdown.classList.remove("is-active", "dropdown");
      tagsDropdown.hidden = true;
    } else {
      tagsDropdown.classList.add("is-active", "dropdown");
      tagsDropdown.hidden = false;
    }
    autoCompleteSuggestions.forEach(function (tag) {
      var tagSuggestion = document.createElement("a");
      suggestionsList.append(tagSuggestion);
      tagSuggestion.setAttribute("href", "javascript:void(0);");
      tagSuggestion.setAttribute("class", "dropdown-item");
      tagSuggestion.setAttribute("tabIndex", 0); // safari blur fix
      tagSuggestion.textContent = tag;

      tagSuggestion.addEventListener("click", function(e){
        event.preventDefault();
        addTag(tag);
        tagsInput.value = "";
        updateDropdownMenu();
      });
    });
  };

  tagsInput.addEventListener("keypress", function(e){
    if (e.key === "," || e.key === "Enter"){
      e.preventDefault();
      var tags = tagsInput.value.split(",");
      var tag = tags.shift();
      tagsInput.value = tags.join(",");
      addTag(tag);
    }
  });
  tagsInput.addEventListener("keyup", function(e){
    updateDropdownMenu();
  });
  tagsInput.addEventListener("focus", function(e){
    focused = true;
    updateDropdownMenu();
  });
  tagsInput.addEventListener("blur", function(e){
    focused = false;
    if (tagsDropdown.contains(e.relatedTarget))
      return;
    tagsDropdown.classList.remove("is-active", "dropdown");
    tagsDropdown.hidden = true;
  });

  updateDropdownMenu();
  return {addTag: addTag, getTags: getTags};
}
