find . -name ".DS_Store" -exec rm -f {} \;

rm -rf out

dpkg-source --commit-using=":" --unapply-patches --skip-patch -b .

chmod 755 debian/*.ex
debuild -us -uc