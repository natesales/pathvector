import React from 'react';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import styles from './index.module.css';
import HomepageFeatures from '../components/HomepageFeatures';
import ImageSwitcher from "../components/ImageSwitcher";

function HomepageHeader() {
    return (
        <header style={{
            textAlign: "center"
        }}>
            <div className="container">
                <ImageSwitcher darkImageSrc={"/img/full-white.svg"} lightImageSrc={"/img/full-black.svg"} width={"50%"}/>
                <div className={styles.buttons}>
                    <Link
                        className="button button--secondary button--lg"
                        to="/docs/about">
                        Learn More
                    </Link>
                </div>
            </div>
        </header>
    );
}

export default function Home() {
    return (
        <Layout
            title={`Pathvector | Edge Routing Platform`}
            description="Description will go into a meta tag in <head />">
            <HomepageHeader/>
            <main>
                <HomepageFeatures/>
            </main>
        </Layout>
    );
}
