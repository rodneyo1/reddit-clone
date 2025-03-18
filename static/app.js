
// Router object: Maps routes to functions
const router = {
    home: async function() {
        let data = await fetchData("home");
        renderPage(data);
    },
};

// Fetch JSON data from Go backend
async function fetchData(route) {
    let response = await fetch(`/api/${route}`);
    return response.json();
}

// Update the page content inside the #app div
function renderPage(data) {
    document.getElementById("app").innerHTML = `
        <h1>${data.title}</h1>
        <p>${data.content}</p>
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