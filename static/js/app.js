var app = angular.module("DogManager", []);

app.controller("DogsCtrl", function($scope, $http) {
  // Array of dog objects
  $scope.dogs = [];
  // Bool true if data fetch fails.
  $scope.error = false;

  // Initialization
  $http.get("/api/dogs")
      .then(function(data, status, headers, config) {
        if (data && data.data) {
          data.data.sort(function(a, b) {
            return a.name > b.name;
          }).sort(function(a, b) {
            return a.adopted > b.adopted;
          });
          for (var i = 0; i < data.data.length; i++) {
            $scope.dogs.push(new Dog(data.data[i]));
          }
          $scope.error = false;
        } else {
          $scope.error = true;
        }
      }, function(data, status, headers, config) {
        $scope.error = true;
      });
});

