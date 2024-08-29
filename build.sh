find . -name ".DS_Store" -exec rm -f {} \;

rm -rf out

chmod 755 debian/*.ex
cp debian/postinst.ex debian/postinst
cp debian/postrm.ex debian/postrm
chmod 755 debian/postinst
chmod 755 debian/postrm

debuild -us -uc