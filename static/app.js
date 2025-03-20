// Router object: Maps routes to functions
const router = {
    home: async function() {
        let data = await fetchData("api/home");
        console.log(data);
        renderPage(data);
    },
    login: async function() {
        renderPage("");
    },

};

// Fetch data from the server using async/await and try/catch
async function fetchData(route) {
    // Check if a route was provided
    if (!route) {
        console.error("No route specified for fetchData");
        return null;
    }
    
    try {
        // Make the fetch request
        const response = await fetch(route);
        
        // Check if the response is OK
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        
        // Log the content type for debugging
        console.log("Content-Type:", response.headers.get("Content-Type"));
        
        // Parse the response as JSON
        const data = await response.json();
        return data;
    } catch (error) {
        console.error("Fetch error:", error);
        
        // Provide more detailed error information for debugging
        if (error instanceof SyntaxError) {
            console.error("JSON parsing failed. The server might be returning HTML instead of JSON.");
        }
        
        return null;
    }
}
class WebPage {
    constructor(data, htmlContent){
        this.data = data;
        this.htmlContent = htmlContent;
    }
}

// Update the page content inside the #app div
function renderPage(WebPage) {
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