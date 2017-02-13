function Dog(raw_data) {
  this.entity_id = raw_data.entity_id;
  this.name = raw_data.name;
  this.gender = raw_data.gender;
  this.breed = raw_data.breed;
  this.age = raw_data.age;
  this.bio = raw_data.bio;
  this.facebook_album_id = raw_data.facebook_album_id;
  this.images = raw_data.images || [];
  this.adopted = raw_data.adopted
  this.nextImage = "";

  function unique(array) {
    return $.grep(array, function(el, index) {
      return index === $.inArray(el, array);
    });
  }

  this.fetchImages = function() {
    var thiz = this;
    $.get('/api/albums?id=' + this.facebook_album_id, null, null, "json")
        .done(function(data) {
          for (var i = 0; i < data.photos.data.length; i++) {
            thiz.images.push(data.photos.data[i].images[0].source);
            thiz.images = unique(thiz.images);
            angular.element($('html')[0]).scope().$apply();
          }
        })
        .fail(function() {
          alert('Failed to fetch images from Facebook.');
        });
  };

  this.addImage = function() {
    if ($.trim(this.nextImage)) {
      this.images.push($.trim(this.nextImage));
      this.images = unique(this.images);
      this.nextImage = "";
    }
  };

  this.deleteImage = function(i) {
    this.images.splice(i, 1);
  };
};
