import React from 'react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import ThemedImage from '@theme/ThemedImage';
import useBaseUrl from '@docusaurus/useBaseUrl';
import HomepageFeatures from '@site/src/components/HomepageFeatures';

import styles from './index.module.css';

function HomepageHeader() {
    return (
        <header style={{
            textAlign: "center",
            marginTop: "10px"
        }}>
            <div className="container">
                <ThemedImage
                    alt="Pathvector Logo"
                    sources={{
                        light: useBaseUrl("/img/full-black.svg"),
                        dark: useBaseUrl("/img/full-white.svg"),
                    }}
                    width={"500px"}
                />

                <p style={{fontSize: "1.25em"}}>
                    Pathvector is a declarative edge routing platform that automates route optimization and control
                    plane configuration with secure and repeatable routing policy.
                </p>
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
            description="Pathvector is a declarative edge routing platform that automates route optimization and control plane configuration with secure and repeatable routing policy.">
            <HomepageHeader/>
            <main>
                <div>
                </div>
                <HomepageFeatures/>
            </main>
        </Layout>
    );
}
