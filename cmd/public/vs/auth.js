

// Register form
const registerForm = document.getElementById("register-form");
if (registerForm) {
    registerForm.onsubmit = async (e) => {
        e.preventDefault();
        const employeeId = document.getElementById("employeeId").value;
        const email = document.getElementById("email").value;
        const mobile = document.getElementById("mobile").value;
        const password = document.getElementById("password").value;
        const confirmPassword = document.getElementById("confirmPassword").value;
        const designation = document.getElementById("designation").value;

        // Validate email format
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) {
            alert("Please enter a valid email address!");
            return;
        }

        // Validate mobile number (simple check for 10 digits)
        const mobileRegex = /^\d{10}$/;
        if (!mobileRegex.test(mobile)) {
            alert("Please enter a valid 10-digit mobile number!");
            return;
        }

        // Validate password strength (at least 8 characters, 1 uppercase, 1 number)
        const passwordRegex = /^(?=.*[A-Z])(?=.*\d).{6,}$/;
        if (!passwordRegex.test(password)) {
            alert("Password must be at least 8 characters long, contain at least 1 uppercase letter, and 1 number!");
            return;
        }

        // Validate password match
        if (password !== confirmPassword) {
            alert("Passwords do not match!");
            return;
        }

        try {
            const res = await fetch("http://localhost:8080/auth/signup", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ 
                    employee_id: employeeId, 
                    email, 
                    mobile, 
                    password, 
                    designation 
                })
            });
            const data = await res.json();
            if (data.message) {
                // Store employeeId in localStorage for OTP verification
                localStorage.setItem("signupEmployeeId", employeeId);
                alert(data.message);
                window.location.href = "/users/otp";
            } else {
                alert("Registration failed: " + (data.error || "Unknown error"));
            }
        } catch (err) {
            alert("Error: " + err.message);
        }
    };
}



// Reset Password form
const resetPasswordForm = document.getElementById("reset-password-form");
if (resetPasswordForm) {
    resetPasswordForm.onsubmit = async (e) => {
        e.preventDefault();
        const employeeId = document.getElementById("employeeId").value;
        const newPassword = document.getElementById("newPassword").value;
        const confirmPassword = document.getElementById("confirmPassword").value;

        try {
            const res = await fetch("http://localhost:8080/users/reset-password", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ employee_id: employeeId, new_password: newPassword, confirm_password: confirmPassword })
            });
            const data = await res.json();
            if (data.message) {
                alert(data.message);
                window.location.href = "/static/login.html";
            } else {
                alert("Password reset failed: " + (data.error || "Unknown error"));
            }
        } catch (err) {
            alert("Error: " + err.message);
        }
    };
}