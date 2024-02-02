
export function getUserInfo() {
    let value = localStorage.getItem("userInfo");
    if (!value) {
        return null;
    }
    return JSON.parse(value);
}

export function setUserInfo(userInfo) {
    if (!userInfo) {
        localStorage.removeItem('userInfo');
        return;
    }

    localStorage.setItem('userInfo', JSON.stringify(userInfo));
    return userInfo;
}