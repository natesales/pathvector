import React from 'react';
import useThemeContext from '@theme/hooks/useThemeContext';

const ImageSwitcher = ({lightImageSrc, darkImageSrc}) => {
    const { isDarkTheme } = useThemeContext();

    return (
        <img src={isDarkTheme ? darkImageSrc : lightImageSrc} alt="Example banner" />
    )
}

export default ImageSwitcher;
