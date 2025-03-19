// Router object: Maps routes to functions
const router = {
    home: async function() {
        let data = await fetchData("api/home");
        console.log(data)
        renderPage(data);
    },
};

// Update the page content inside the #app div
function fetchData(route) {
    return fetch(route)
        .then(response => {
            if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
            return response.json();
        })
        .catch(error => {
            console.error("Fetch error:", error);
            return null;
        });
}

// Update the page content inside the #app div
function renderPage(data) {
    document.getElementById("app").innerHTML = `
        <h1>${data.posts}</h1>
        <p>${data.IsLoggedIn}</p>
    `;
}

// Navigate function: Calls the correct handler
function navigate(route) {
    if (router[route]) {
        router[route]();
        window.history.pushState({}, "", `/${route}`); // Update URL
    } else {
        console.error("Route not found:", route);
    }
}

// Handle browser back/forward buttons
window.onpopstate = () => {
    let path = window.location.pathname.substring(1);
    navigate(path || "home");
};

// Load the default page when the app starts
window.onload = () => {
    let path = window.location.pathname.substring(1); // allow refreshing on any page
    navigate(path || "home");
};

// Call when page loads
document.addEventListener('DOMContentLoaded', fetchData);