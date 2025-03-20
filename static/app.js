// Router object: Maps routes to functions
class WebPage {
    constructor(data, htmlContent){
        this.data = data;
        this.htmlContent = htmlContent;
    }
}
const router = {
    home: async function() {
        let data = await fetchData("api/home");
        // console.log(data);
        renderPage(data);
    },
    login: async function() {
        renderPage(new WebPage(null, `
        <div class="auth-container">
        <h1>Login</h1>

        <!-- Google Sign-In Button -->
        <a href="/auth/google/login" class="google-btn">
            <img src="/src/google.jpeg" alt="Google Logo">
            <span>Sign in with Google</span>
        </a>

        <!-- GitHub Sign-In Button -->
        <a href="/auth/github/login" class="github-btn">
            <img src="https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png" alt="GitHub Logo">
            <span>Sign in with GitHub</span>
        </a>

        <div class="oauth-divider">aa
            <span>or</span>
        </div>

        <!-- Traditional Login Form -->
        <form method="POST" action="/login">
            <label for="email">Email:</label>
            <input type="email" id="email" name="email" placeholder="example@gmail.com" required>
            <br>
            <label for="password">Password:</label>
            <input type="password" id="password" name="password" required>
            <br>
            <button type="submit">Login</button>
        </form>
        <p>Don't have an account? <a href="/register">Register here</a></p>
        <p class="home-link"><a href="/">← Back to Homepage</a></p>
        </div>   
        `));

        document.getElementById("loginForm").addEventListener("submit", handleLogin);
        // renderPage("");
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


// Update the page content inside the #app div
function renderPage(WebPage) {
    const app = document.getElementById("app");
    
    // Create a wrapper element and set content
    const container = document.createElement("div");
    container.innerHTML = WebPage.htmlContent;

    // Replace existing children without modifying `#app` itself
    app.replaceChildren(container);
}



async function handleLogin(event) {
    event.preventDefault();

    const username = document.getElementById("username").value;
    const password = document.getElementById("password").value;

    try {
        const response = await fetch("login", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ username, password }),
        });

        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();

        if (data.success) {
            navigate("home");
        } else {
            alert("Login failed: " + data.message);
        }
    } catch (error) {
        console.error("Login error:", error);
        alert("Login failed. Please try again.");
    }
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

let button=document.getElementById("login-button")
button.addEventListener("click", (event) => {
    if (event.target.id === "login-button") {
        event.preventDefault(); // Prevents the default anchor tag behavior
        navigate("login"); // Navigate using SPA logic
    }
});
