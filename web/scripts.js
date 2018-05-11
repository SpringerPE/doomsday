
    // /v1/cache with basic auth
    // then pass token
    $(document).ready(function () {
      console.log("document ready");
      document.cookie = "username=John Doe";
      generateCards();
    });

    $("#search-box").change(function () {
      alert("Handler for .change on searchbox called.");
      searchFilter();
    });

    //search by CN
    function searchFilter() {
      var input, filter, ul, li, a, i;
      input = document.getElementById("searchBox");
      filter = input.value.toUpperCase();
      ul = document.getElementById("card-list");
      li = ul.getElementsByClassName("card-header");
      // Loop through all list items, and hide those who don't match the search query
      for (i = 0; i < li.length; i++) {
        a = li[i].getElementsByClassName("card-header")[0];
        if (a.innerHTML.toUpperCase().indexOf(filter) > -1) {
          li[i].style.display = "";
        } else {
          li[i].style.display = "none";
        }
      }
    }
    // [
    //     {
    //         "path": "/api/v0/deployed/products/cf-32403a409e48e697b084/credentials/.mysql.backup_server_certificate",
    //         "common_name": "streaming-mysql-backup-tool",
    //         "not_after": 1583347595
    //     },
    //     {
    //         "path": "/api/v0/deployed/products/cf-32403a409e48e697b084/credentials/.diego_database.silk_daemon_client_cert",
    //         "common_name": "silk_daemon_client_cert",
    //         "not_after": 1583347595
    //     },
    //     {
    //         "path": "/api/v0/deployed/products/pivotal-container-service-4fcd739a058e0c206566/credentials/.pivotal-container-service.pks_tls",
    //         "common_name": "*.pks.system.pcf20.starkandwayne.com",
    //         "not_after": 1584213813
    //     }
    // ]
    // when you make the request you might get a 401
    // if you get a 401 you should then ask for password
    function generateCards() {
      $.get("http://localhost:8111/v1/cache", function (data) {
        console.log(data);
        for (var i = 0; i < data.length; i++) {
          console.log(data[i])
        $(".flex-container").append(
          '<li><div class="card dropshadow"><div class="card-header">' +
          data[i].common_name +
          '</div><div class="card-main"><i class="material-icons">fingerprints</i><div class="main-description">' +
            unixTimeToJSDate(data[i].not_after).toLocaleDateString() +
          "</div></div></div></li>"
        );
        }
      }).fail(function () {
        alert("woops"); // or whatever
      });
    }

    function grabInfoFromAPI() {
      // Assign handlers immediately after making the request,
      // and remember the jqxhr object for this request
      $.get("http://localhost:8111/v1/cache", function (data) {
        console.log(data);
        return data
      }).fail(function () {
        alert("woops"); // or whatever
      });
    }

    function unixTimeToJSDate(unixTime) {
      // Create a new JavaScript Date object based on the timestamp
      return new Date(unixTime * 1000);
    }

    // api({
    //     type: "GET",
    //     url: "/v2/auth/logout",
    //     success: function () {
    //       document.location.href = '/';
    //     },
    //     error: function (xhr) {
    //       if (xhr.status >= 500) {
    //         $('#viewport').html(template('BOOM'));
    //       } else {
    //         document.location.href = '/';
    //       }
    //     }
    //   })