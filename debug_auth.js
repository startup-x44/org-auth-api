// Add this to your browser console to debug auth state
console.log('Current auth state:', JSON.parse(localStorage.getItem('auth-storage') || '{}'))
console.log('Access token:', localStorage.getItem('access_token'))
console.log('User data:', JSON.parse(localStorage.getItem('user_global') || 'null'))