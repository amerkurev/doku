export function getLocalStorage(key, initialValue) {
  try {
    // Get from local storage by key
    const item = window.localStorage.getItem(key);
    // Parse stored json or if none return initialValue
    return item ? JSON.parse(item) : initialValue;
  } catch (error) {
    return initialValue;
  }
}

export function setLocalStorage(key, value) {
  try {
    // Save to local storage
    window.localStorage.setItem(key, JSON.stringify(value));
  } catch (error) {
    console.error(error);
  }
}
